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

**A graph may not hold a thing together with a device linked to it at the same
time.** The validity of the thing node, the validity of the device node and the
period of the link must have no instant in common; a node or a link without
bounds is valid always, so for graphs and links without validities the rule is
the old one: any link, ended ones included, forbids both in one graph. The
device's values stand for the thing during the link, so both nodes live during
that period count it twice. Because the validities are compared, a device node can
be replaced by a thing node from a date on: the device node ends where the thing
node starts (see [graph-validity.md](graph-validity.md)). The rule is checked from
both sides, `PUT /graphs` and link creation, each against the other's
validities.

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

**A thing is a consumer or a meter.** `kind` is empty or `consumer` for the
behaviour described above, `meter` for a thing that stands for a metering point.
An empty kind is a consumer, so stored things need no migration. Only the links
of a meter carry `readings` (`service_id`, `column`, `from_value`, `until_value`:
the counter reading at installation and at removal) and `factor` (transformer
ratio, greater than 0, absent means 1); on a consumer they are a 400. Link
creation takes `readings` and `factor`, ending a link takes `until_values`, a
list of `{service_id, column, until_value}` that must each name an existing
reading of the link. `PUT /things/{id}` without `kind` keeps the stored one, like `device_links`; it can change `kind`, but not from `meter`
to `consumer` while a link has readings or a factor. The graph rule above holds
for both kinds.

**Swapping a link is one write.** `POST /things/{id}/device-links/{link_id}/swap`, body
`{"at", "until_values", "device_id", "readings", "factor"}`: the open link ends at `at` (with `until_values`, as when
ending) and a link to `device_id` starts at `at` (with `readings` and `factor`, as when creating), in the same write of
the thing, so a refusal of either half leaves the thing as it was. `device_id` is the same device for a counter reset or
a changed factor, another one for a replaced device or meter. `at` defaults to now, may lie in the past, and must be
after the start of the ended link (400 naming both instants); the link must be open. The new link follows the rules of
creation, checked against the thing as it will be: a link of that device from `at` on, on any thing, is an overlap
(400), the same device is not blocked by the link that ends, and a graph holding the thing and the device then is
refused. `readings` may not carry an `until_value`. Needs write on the thing and read on the device. Graphs are not
touched. The device lookup the Web-UI needs to find the open link of a device is `GET /things?device_ids=<id>`: it lists
things with an open or ended link to the device, the link is the one of the answer with the device and no `until`.

## Limits

**Writes of one thing are compare-and-set.** Every stored thing has a `revision` (the time of its last write in
nanoseconds, not part of the model, missing in documents from before it existed). `PUT /things`, link creation, link
end, swap and the ending of links on a device delete read the thing, compute the new document and write it only if
the revision is still the one that was read; otherwise they run again on what is stored now, up to five times, then the
answer is 409. A swap and a rename, or two swaps of different links, therefore never overwrite each other; two swaps of
the same link end with the second one refused (the link is already ended).

The overlap check and the graph rule stay checks before the write and are not atomic across documents: two concurrent
link creations for the same device on two different things can both pass, because each protects only its own thing.
Links are rare and mostly created by hand, so this was accepted. Likewise the graph rule is checked against the graphs
as they are when the link is written, not held against a concurrent graph write.

The controller evaluates links and nodes itself and uses the database filters
(`ThingListOptions.DeviceIds`, `GraphListOptions.ThingIds`) only to narrow the
candidates. That keeps the checks correct against the test database, which
ignores list options, and against mongo, which returns every match when
`Limit` is 0 (see [testdb-and-mongo-disagree-on-limit-zero.md](testdb-and-mongo-disagree-on-limit-zero.md)).
