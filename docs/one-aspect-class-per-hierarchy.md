# An aspect hierarchy has one aspect-class, assigned by its root

## Applies when

Writing aspects that carry an `aspect_class_id`, reading it back off an aspect or
an aspect-node, deleting an aspect-class, or putting more than one aspect on a
content variable. Follows from the validation in `lib/controller/aspects.go`,
`lib/controller/aspectclasses.go` and `lib/controller/variable.go`.

**Not this if**: the question is which aspect-classes exist and how to manage
them. That is plain CRUD on `/aspect-classes`, documented in the swagger, and
none of the rules below apply to the aspect-class itself — they all constrain who
may *reference* one.

## Only the root assigns the class, everything below inherits it

`ResolveAspectClassIds` runs before validation on every aspect write and gives
each aspect of the tree the class of its root. Three consequences:

- A sub-aspect that carries **no** class gets the root's — including nested ones,
  at any depth.
- A sub-aspect may **repeat** the root's value. That is a no-op, not an error.
- A sub-aspect that carries a **different** value is rejected with 400, and so is
  a sub-aspect that carries one while the root has none. Classifying a subtree
  from below is not possible; the assignment belongs to the root.

The stored tree is therefore uniform: either every aspect carries the same class,
or none does. A referenced class has to exist — an unknown id is a 400.

The resolution also runs in `ValidateAspect`, because the `?dry-run=true`
endpoint and the import path never pass `SetAspect`, and in the mgw mirror, which
writes aspects straight to the database.

## Why only the root assigns it

Two other readings of "a hierarchy has one class" were on the table and rejected:

- **Allow a partially classified tree.** Literal inheritance leaves the root
  empty when nothing is assigned there, so a sub-aspect could classify its own
  subtree and the tree would still hold only one distinct class. Rejected because
  a tree whose upper levels are unclassified makes "the class of this aspect"
  depend on where you enter the tree, and the delete check would then have to
  distinguish carried from inherited.
- **Let the class propagate upward.** Whatever single class appears anywhere gets
  applied to the whole tree, root included. Rejected because assignment would
  then have no single owner: two people classifying two different subtrees would
  each believe they set the hierarchy's class, and the second write would win
  silently.

What is left is the strictest of the three and the only one that has an owner:
the root. The cost is that a subtree cannot be classified on its own — moving it
out into a hierarchy of its own is the way to give it a different class.

## Assignment is optional by default

`aspect_class_id_required` in the config, default `false`, decides whether an
aspect must resolve to a class. It is checked against the root, which through the
inheritance settles it for the whole tree. With the flag off an entirely
unclassified hierarchy is valid — which is why the existing tests need no
aspect-classes at all.

## The aspect-nodes are the queryable index

`CreateAspectNodes` copies the class into every node. An aspect tree is stored as
**one** nested document, so the aspect collection cannot answer "which aspects use
this class" at arbitrary depth; the node collection has one flat document per
aspect and does. Both listings take the filter:

```
GET /v2/aspects?aspect_class_ids=<a>,<b>       # the hierarchies, so only roots
GET /v2/aspect-nodes?aspect_class_ids=<a>,<b>  # every single aspect
```

An empty `aspect_class_ids=` returns nothing; leaving the parameter out filters
nothing. Both collections carry an index on the field.

That difference is easy to trip over: filtering `aspects` by a class returns one
element per tree, filtering `aspect-nodes` returns one per aspect. For a
four-aspect hierarchy that is 1 versus 4.

## An aspect-class cannot be deleted while an aspect carries it

`DELETE /aspect-classes/{id}`, and its `?dry-run=true` form, refuse with 400 and
name the offending aspects:

```
still in use by 4 aspect(s): air (urn:infai:ses:aspect:air), inside_air (…), morning_air (…), outside_air (…)
```

The count is exact; the list of names is capped at `aspectClassUsageErrorLimit`
and says "and N more" beyond that. Deleting the aspects releases the class.

## A content variable carries at most one aspect per class

`ValidateVariable` resolves every entry of `ContentVariable.AspectIds` to its
node and refuses a device-type whose variable holds two aspects of the same
non-empty class. Two **unclassified** aspects do not collide, because there is no
class to collide on. The same aspect listed twice is refused with a message of
its own rather than being reported as two aspects.

For a client the useful phrasing is not about classes at all. Since a hierarchy
carries exactly one class, the rule means: **at most one aspect out of any
classified hierarchy may sit on a content variable.** `inside_air` and
`outside_air` together were valid until the hierarchy above them got a class;
afterwards they are a 400. Aspects from *different* hierarchies remain fine, and
so do two aspects of one hierarchy as long as that hierarchy has no class — which
is the boundary worth knowing, because it means classifying an existing hierarchy
can invalidate device-types that were valid before.

**This is only checked when the device-type is written.** Changing an aspect's
class later can leave an already stored content variable in violation, and
nothing notices. Catching that would mean scanning the device-type criteria on
every aspect write, the way `setDeviceTypeSyncHandler` scans device-groups —
see [not-every-device-type-read-passes-the-controller.md](not-every-device-type-read-passes-the-controller.md)
for what that shape of problem costs.
