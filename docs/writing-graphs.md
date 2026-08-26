# Writing graphs

## Applies when

Creating or updating graphs through `PUT /graphs` / the Go client's `SetGraph`,
as a service that authors graph content. Both rules confirmed 2026-08-26
against a running repository (`device-repository:dev`, client v0.3.2) while
building a second graph author.

**Not this if**: a 403 on a graph id that **does exist** — that is a real
permission problem of the token, not this note's case.

## A new graph's id can only come from the server

`SetGraph` with a self-invented, not-yet-existing id is answered with **403**,
not 404: a PUT to an unknown graph id runs a permission check against a
resource that has no permissions yet, and for a non-admin token that check
fails. It looks exactly like a token or role problem and is neither — create
with an **empty id** (the repository assigns one and returns it) and store the
returned id for later upserts.

The same 403 shows up when a stored graph was deleted by hand: its permissions
went with it, so an upsert to the remembered id can never succeed again. A
writer that remembers ids has to decide what to do then; auto-recreating on
403 duplicates graphs whenever a 403 is transient.

## Valid() constrains edge weights, and the server enforces it

`SetGraph` validates and answers 400 on failure. Every edge needs
`0 < Weight <= 100`, and each node's **outgoing weights must sum to 0 or 100**.
A "neutral" weight of 0 is therefore rejected; for a tree, where every node
has exactly one parent edge, the only valid value is 100 — which is also the
semantically correct one (a node passes its whole flow to its one parent).

Duplicate node ids and duplicate device `resource_id`s are rejected as well,
so a writer whose sources can point two entities at one device has to
deduplicate before sending.
