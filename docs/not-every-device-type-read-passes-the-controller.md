# Not every device-type read passes the controller

## Applies when

Changing how a `models.DeviceType` field is interpreted — a new field, a
deprecated one, a normalisation applied at the controller boundary. Follows from
reading the call sites in `lib/controller` and `lib/mgwmirror` and from a test
that reproduces the data loss described below, not from a single observation.

**Not this if**: the concern is the `deviceTypeCriteria` collection. Those rows
carry a copy of the interpreted fields too, but they are only ever written by
`createCriteriaListFromDeviceType` inside `db.SetDeviceType`, so every path that
produces them already sits behind the write normalisation. Checking that
collection and concluding "no risk" is the false trail — it is the one derived
store that is safe.

## Two paths read or write device-types without a controller read

Most device-type reads funnel through `Controller.readDeviceType` or
`modifyDeviceTypeList`, so a normalisation placed there covers them. Two do not:

- `getDeviceGroupCriteriaOfDevice` in `lib/controller/devicegroupgenerated.go`
  calls `this.db.GetDeviceType` directly and walks the content variables itself.
- `lib/mgwmirror/source.go` pulls resources from a remote device-repository over
  the client and writes them with `db.Set*`, so the controller's **write**
  normalisation never runs on them either.

Both are easy to miss because the surrounding code lives in `lib/controller` and
looks like it is on the inside of the boundary.

## Why the device-group path is the expensive one

`getDeviceGroupCriteriaOfDevice` does not just read. Its result is persisted:

```
UpdateDeviceGroupCriteria(dg)
  -> GetDeviceGroupCriteria(dg.DeviceIds)
       -> getDeviceGroupCriteriaOfDevice(device)   // raw db.GetDeviceType
  -> setDeviceGroup(dg, user)                      // overwrites dg.Criteria
```

and it runs on two ordinary events:

- `setDeviceTypeSyncHandler` — on **every** device-type write, for every
  device-group of every device of that device-type
  (`lib/controller/devicetype.go`)
- `EnsureGeneratedDeviceGroup` via `setDeviceSyncHandler` — on device create and
  update (`lib/controller/device.go`)

So a field interpretation that only holds after a controller read degrades
silently: the stored device-groups are rewritten from the raw device-type, and
whatever the raw record does not carry is gone from the persisted criteria. Not
a display artefact — the documents change.

This is what happened while `ContentVariable.AspectId` was replaced by
`AspectIds`. Records written before the change carry only `AspectId`. The read
path filled `AspectId` from `AspectIds` but not the reverse, so on the raw path
`AspectIds` was empty and the generated criteria came out with an empty aspect:

```
[{event urn:infai:ses:measuring-function:getTemperature  } {request urn:infai:ses:measuring-function:getTemperature  }]
```

Touching one device-type after deploying would have stripped the aspect from the
device-groups of all its devices.

## What to do instead

Make the compatibility normalisation symmetric and apply it on the raw paths too:

- On read, fill the new representation from the old one **before** deriving the
  old one back from the new. `syncContentVariableAspects` in
  `lib/controller/contentvariableaspects.go` is the worked example.
- Call the write normalisation right after every raw `db.GetDeviceType` that
  interprets the field, and before every raw `db.Set*` that persists one. That is
  why `SetContentVariableAspectIdsOnWrite` is exported.

`ValidateDeviceType` needs it as well, for a different reason: the `?dry-run=true`
endpoint and the `/invalid` listing validate device-types that never went through
a write, so without it the check silently passes on unmigrated records.

## How to catch it in a test

Writing a legacy record through the API is impossible once the write path
normalises. Build the database layer directly instead — that is the only way to
produce the pre-migration shape:

```go
db, err := mongo.New(config)
defer db.Disconnect()
err = db.SetDeviceType(context.Background(), legacy, func(models.DeviceType) error { return nil })
```

then assert through the client that the read returns both representations and
that the generated device-group still carries the aspect. See
`lib/tests/content_variable_aspect_ids_test.go`, subtest
`device-type stored with aspect_id only`.
