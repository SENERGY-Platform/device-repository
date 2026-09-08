# A device-group may hold modified device ids

## Applies when

Reading or deriving anything from `models.DeviceGroup.DeviceIds` — the generated
criteria, a listing filter, a membership check — or writing a new endpoint that
takes device ids from a request body. A group member may be a **modified** device
id such as `urn:infai:ses:device:1$service_group_selection=sg1`, and that id is not
in the `devices` collection.

**Not this if**: the modified id in hand is a **device-type** id. Those are handled
by `IncludeModified` on the listing options and by `modifyDeviceTypeList`, they do
exist as rows in the `deviceTypeCriteria` collection, and none of the splitting
below applies to them. Mistaking the two is the easy error, because both spell the
modifier the same way and `SplitModifier` takes either.

Also not this if the question is whether a device-type read is normalised at all
before it is interpreted — that is
[not-every-device-type-read-passes-the-controller.md](not-every-device-type-read-passes-the-controller.md).

## A modified device is derived, never stored

`idmodifier.SplitModifier` cuts an id at `$` into the pure id and the decoded
modifier. Only the pure id is stored; the modified one is produced on read by
`Controller.modifyDevice`, which applies the modifier to the device and appends it
to the device's `DeviceTypeId` as well. `ServiceGroupSelectionIdModifier` is the
only modifier there is today: it keeps the services of one service-group plus the
ones belonging to no group.

Three consequences, and the first two are what breaks:

- **Every database lookup needs the pure id.** `db.GetDevice`, `db.GetDeviceType`
  and `db.ListDevices` know nothing about `$`. Handed a modified id they return
  "not found" — and `db.GetDeviceType` returns a zero-valued device type next to
  its `exists` flag, so a caller that ignores the flag silently proceeds with a
  device type that has no services and derives no criteria at all.
- **A modified device answers fewer services**, so its criteria are not the stored
  device's. Whatever is derived from a group member has to apply the modifier
  first, or a group of `device$sg1` gets the criteria of every service group at
  once.
- **A permission check may take the raw id.** permissions-v2 splits the modifier
  itself and checks the pure resource, in `CheckPermission` and in
  `CheckMultiplePermissions` alike, so passing the modified id through is correct
  and passing the pure one loses nothing.

## The shape that works

`Controller.ListDevices` and `Controller.ListExtendedDevices` already do this, and
it is worth copying rather than reinventing: collect the pure ids, keep a
`pureId -> []rawId` map, query the database with the pure ids, then expand each
result back into one entry per raw id and apply that id's modifier.

`GetDeviceGroupCriteria` had none of this until 2026-09-08 and returned an empty
criteria list for any group holding a modified member, without an error anywhere.

## Why a missing device-type is not an error here

`getDeviceGroupCriteriaOfDevice` logs a warning and yields no criteria when the
device-type does not exist, rather than failing. A device can outlive its
device-type in stored data, and the same function runs inside a startup migration
over the whole database — see
[writing-a-startup-migration.md](writing-a-startup-migration.md).
