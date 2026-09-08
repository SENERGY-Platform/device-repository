# Not every controller method checks permissions

## Applies when

Exposing an existing `Controller` method through a new HTTP endpoint, or reusing
one from a handler that already has the caller's token. Some methods take a token
and check it, some take no token and read straight from `this.db`. The signature is
the only signal, and it is easy to read the wrong way round.

**Not this if**: the method takes a `token string` and a `model.AuthAction`, like
`ReadDevice`, `ReadExtendedDevice`, `ListDevices` or `ListExtendedDevices`. Those
call `permissionsV2Client` themselves and need nothing from the handler. Sampling
one of those and concluding "the controller checks" is the false trail — it is true
for the read and listing surface and false for the helpers underneath it.

## A method without a token has no way to check

`GetDeviceGroupCriteria(deviceIds []string)` is the example. It takes no token
because it was written for callers that have no user: `UpdateDeviceGroupCriteria`
in a sync handler, and `runGeneratedDeviceGroupCriteriaMigration` at startup. Both
are right to skip the check. It reads devices and device-types out of `this.db`
directly.

Put such a method behind a handler unchanged and the endpoint answers for any id
the caller can name. For the criteria that is not an abstract leak: the answer
enumerates the functions, aspects and device-class a device serves, so guessing an
id tells you what someone else's device measures and controls.

`POST /device-group-helper` therefore checks the ids of its request body against
the device topic before it derives anything, and answers `403` for the first id the
caller may not read. The check sits in the controller method that the endpoint
calls, not in the handler, so it cannot be forgotten by a second caller.

## The rule

A method that reads `this.db` and takes no token is an internal method. Exposing it
means adding the check — at the boundary the endpoint calls, with the caller's
token — not documenting that the endpoint is internal. There is no internal port
here; everything on the router is reachable through the gateway.
