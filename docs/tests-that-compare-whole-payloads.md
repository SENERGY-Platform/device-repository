# Which tests break when a response gains a field

## Applies when

Adding a field to anything that appears in an API response — a new field on a
`models.*` struct from the models module, or an existing one the service starts
populating. Observed while `ContentVariable.AspectIds` began to be written;
the affected packages below are the complete list from that run.

**Not this if**: the test checks individual fields, like
`result.Services[0].Inputs[0].ContentVariable.AspectId != a1Id` in
`lib/tests/manager_legacy/devicetype_test.go`. Those pass unchanged. Most tests
in the repository are of that kind, so sampling a few and concluding "an added
field is harmless" is exactly the wrong conclusion — the packages below compare
the whole serialized document and fail on any addition.

## The five packages

| Package | What it compares |
|---|---|
| `lib/tests/repo_legacy` | `TestConfigurables` — Go literals for `[]model.DeviceTypeSelectable` |
| `lib/tests/repo_legacy/aspect_hierarchy_test` | `TestDeviceTypeFilterCriteria` — device-type literal, compared as JSON maps |
| `lib/tests/repo_legacy/devicetypeselectables_test` | Go literals **and** one long inline JSON string **and** the `a5t_testcase/expected*.json` fixtures |
| `lib/tests/repo_legacy/devicetypes_test` | `models.DeviceType` literals via `reflect.DeepEqual` |
| `lib/tests/semantic_legacy` | `TestDeviceType` — expected values as marshalled JSON strings |

`lib/tests`, `lib/tests/manager_legacy`, `lib/controller` and
`lib/database/mongo` were unaffected.

## Expected values live in three shapes

Updating them is mechanical but the shapes differ, and a blind
search-and-replace hits the wrong ones:

- **Go literals** inside `models.ContentVariable{...}`. In the same files,
  `model.FilterCriteria` and `models.DeviceGroupFilterCriteria` literals carry a
  field of the same name that must **not** be touched. Distinguish by the
  enclosing composite literal, not by the field name.
- **Inline JSON strings** assigned to a variable in the test body. Key order
  follows the Go struct field order, because the fixture was produced by
  marshalling.
- **JSON fixtures** under
  `lib/tests/repo_legacy/devicetypeselectables_test/a5t_testcase/`. `expected.json`
  is minified, `expected_with_modified.json` and
  `expected_service_must_match_short.json` are pretty-printed — a regex written
  for one form silently matches nothing in the other.

Watch for empty values: a variable with an empty deprecated field must not gain
an empty entry in the list form. That mistake produced `"aspect_ids": [""]` and
one failing subtest.

## Proving the diff is only the new field

Before editing ~90 expected values it is worth knowing that nothing else moved.
Temporarily strip the new field again on the read path, run the failing packages,
and confirm they go green. If they do, every difference was the added field and
the expectation update is safe. That check cost one test run and removed the
guesswork.
