# Things and device links

## Applies when

Changing or extending the thing resource (`lib/controller/things.go`,
`models.Thing` in `SENERGY-Platform/models`), the device link endpoints, or the
graph rule that couples both. Written 2026-11-12 with v3.10.0 to v3.10.4.

**Not this if**: the question is how to call the endpoints. That is in the
swagger docs; this note records why they behave the way they do.

## What a thing is

A thing is a consumer without measurements of its own, for example an appliance
known only from its nameplate. It carries `attributes` (power, relative usage,
whatever its creator knows) and `content_variables`. The service stores the
content variables and generates missing ids, but does **not** compute values
from the attributes: the side that evaluates the data does, because only it knows
the period and the method. A device can measure a thing for a period; during
that period the device's values take precedence over the thing's content
variables, again on the evaluating side.

## Decisions

**Links have their own endpoints.** `POST /things/{id}/device-links` creates one,
`POST /things/{id}/device-links/{link_id}/end` ends it. `PUT /things/{id}` keeps
the stored links and ignores `device_links` in the body. Links in the thing
document with a check against the stored history would have meant validating
every change to the history; the two endpoints make "never deleted, only ended"
hold by construction.

**Periods are half-open, `[from, until)`, `until: null` is open.** Two periods
that touch, one ending exactly when the next starts, do not overlap, so a device
moved from one thing to the next has no gap and no double second.

**Periods of one device never overlap, across all things.** A device measures one
thing at a time. A thing may be measured by several devices at once.

**Missing `from` or `until` means now, cut to milliseconds in UTC.** Mongo stores
dates with millisecond precision; without the cut the in-memory test database and
mongo would answer the same request differently. Ending a link that starts now
or in the future without an `until` sets `until = from + 1ms`; the same rule
applies when a device is deleted and its open links are ended.

**A graph may not hold a thing together with a device ever linked to it.** Ended
links count as well. The device's values stand for the thing during the link, so
both in one graph count that period twice, and a historical evaluation reaches
back into ended links. The rule is checked from both sides, `PUT /graphs` and
link creation.

**A thing may not be a graph's end node**, the same rule as for devices: the end
node is the aggregate on top, not a single consumer.

**A non-admin creating a thing always becomes its owner.** `SetGraph` takes a
given `owner_id` for a new graph unchecked; for things a foreign `owner_id` on
create is replaced by the caller, so nobody can create things in someone else's
name.

**Linking needs write on the thing and read on the device.** Nothing changes on
the device, and anyone who can read it can read its data anyway.

**Deleting a thing removes it from all graphs; deleting a device ends its open
links.** The thing is deleted with its links. The device's links stay as history
pointing to a device that no longer exists.

## Limits

The overlap check and the graph rule are not atomic; the repository uses no
mongo transactions. Two concurrent link creations for the same device on two
things can both pass, and a `PUT /things/{id}` running concurrently with a link
creation can drop that link, because both write the whole document. Links are
rare and mostly created by hand, so this was accepted.

The controller evaluates links and nodes itself and uses the database filters
(`ThingListOptions.DeviceIds`, `GraphListOptions.ThingIds`) only to narrow the
candidates. That keeps the checks correct against the test database, which
ignores list options, and against mongo, which returns every match when
`Limit` is 0 (see [testdb-and-mongo-disagree-on-limit-zero.md](testdb-and-mongo-disagree-on-limit-zero.md)).
