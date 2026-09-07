# Listing devices omits the display name

## Applies when

Reading a list of devices from this service and needing the name a user
recognises — `GET /devices`, `GET /extended-devices`, `GET /local-devices`.

**Not this if**: the question is how `display_name` comes about or how to
change it. That is `docs/display-name-is-derived.md`. This note is only about
which listing endpoint returns it.

## GET /devices has no display_name at all

`GET /devices` returns `models.Device`, and that struct carries no
`display_name` field — the derived value lives on `models.ExtendedDevice`
(`lib/model/connectionstate.go`), which only the extended endpoint builds
(`lib/controller/device.go`, `getDeviceDisplayName`).

The consequence is easy to miss because nothing is missing from the response:
`name` is there and looks like an answer. For a device created by hand it even
is one. For a device an import created, `name` is the identifier that source
produced — a wmbus device is called something like

```
(IOM) IOmeter, Germany (0x25ed) Radio converter (meter side) (0x37) encrypted 00000541
```

while the web ui shows the nickname a user gave it. A caller that lists devices
and prints `name` therefore shows something nobody recognises, with no error to
suggest the wrong endpoint was used, and the more devices came from imports the
worse it reads.

## Use /extended-devices, and fulldt when services are needed

`GET /extended-devices` takes the same query parameters as `GET /devices`
(`lib/api/devices_extended.go`) and adds `display_name`, `device_type_name`,
`connection_state`, `permissions` and `shared`.

It also takes **`fulldt`**, which inlines the whole device type — services with
their outputs and content variables included. A caller that needs the services
of every listed device therefore has a choice between one request and one
request per distinct device type:

```
GET /extended-devices?limit=1000&offset=0&fulldt=true
```

For a set of devices spread over a few dozen types, that is the difference
between one call and a few dozen.

## Paging

Both endpoints take `limit` and `offset`, and **`limit` defaults to 100 when it
is omitted** — a caller that leaves it out and reads the response as the whole
list silently sees the first hundred devices. Setting `ids` ignores `limit` and
`offset` entirely.

The handlers pass the limit through without capping it, so asking for more than
the account holds returns everything in one page today. A loop that wants to
stay correct if that ever changes should end on an **empty** page and advance
the offset by the number of rows actually returned, rather than treating a page
shorter than the requested limit as the end.
