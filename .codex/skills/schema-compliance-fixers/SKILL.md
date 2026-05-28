---
name: schema-compliance-fixers
description: Use when adding or modifying schema_compliance fixer functions in this Go repo, including stage selection, schema-loss gating, tests, and validation commands.
---

# Schema Compliance Fixers

Use this skill when adding a new fixer to `schema_compliance`.

## First Steps

- Read `schema_compliance/schema_compliance.go`, `schema_compliance/fixes_list.go`, and the closest existing fixer/test pair before editing.
- Use test-first development: add focused tests, run `go test ./schema_compliance` and confirm the expected failure, then implement.
- Keep fixers private. The public API should remain `Ensure`, `ValidateJSON`, `ValidateAgainstSchema`, and the exported error types unless the user explicitly asks otherwise.

## The Three Stages

`Ensure` runs these stages in order:

1. `oneTimeFixes()` runs once before JSON syntax repair.
   - Use for extracting an actual JSON-ish payload from surrounding LLM text.
   - Example: fenced JSON extraction from prose.
   - The fixer type is `fixFunc func(string) (string, bool)`.

2. `jsonSyntaxFixes()` runs repeatedly until no fixer changes the string, but only before strict JSON validation.
   - Use for making invalid JSON become valid JSON, without considering the schema.
   - Examples: relaxed JSON parsing, missing final object/array closers from truncation.
   - Be conservative: consume the whole input, avoid guessing missing semantic values, and return unchanged if the result is not clearly valid JSON.
   - The fixer type is `fixFunc func(string) (string, bool)`.

3. `schemaComplianceFixes()` runs repeatedly after JSON is valid but before final schema validation.
   - Use for transforming valid JSON into a closer match for the compiled schema.
   - Examples: unwrap response envelopes, unwrap single-item arrays, move fields between nesting levels, scalar coercion.
   - Every schema-stage fixer must use `schemaLoss(candidate, schema) < schemaLoss(current, schema)` before accepting a candidate.
   - The fixer type is `schemaFixFunc func(string, *jsonschema.Schema) (string, bool)`.

## Schema-Stage Patterns

- Parse once with `parseAndNormalizeJSON`; return unchanged if parsing fails.
- Generate canonical candidates with `marshalCanonicalJSON`.
- Reuse existing helpers when possible:
  - `candidateSchemaBranches` for `Ref`, `OneOf`, `AnyOf`, and `AllOf`.
  - `sortedObjectKeys` and `sortedSchemaPropertyNames` for deterministic traversal.
  - `cloneJSONValue`, `cloneJSONObject`, and `cloneJSONArray` before mutating candidate values.
  - `arrayItemSchemas` and `itemSchemasForIndex` when traversing arrays.
- Make only one logical change per fixer invocation. The schema-stage loop will call fixers repeatedly when each change improves loss.
- Do not overwrite existing fields, drop unrelated data, or delete fields unless that specific fixer is explicitly designed and tested to do so.
- If multiple candidates exist, choose the first deterministic candidate that reduces global loss.

## Loss Function Guidance

`schemaLoss` is a deterministic heuristic used only to decide whether a repair is an improvement. Final acceptance still depends on `schema.Validate`.

Adjust `schemaLoss` only when a fixer needs to distinguish two candidates that current scoring treats incorrectly. Add internal tests showing the desired ordering before changing it.

Safe changes:

- Add scoring for a JSON Schema feature already used by a fixer, such as formats, branch selection, array item schemas, required fields, or enum/const mismatches.
- Tune weights when a more schema-relevant near miss should score lower than a less relevant shape, with tests proving the ordering.
- Keep unsupported complex features on a fixed fallback penalty rather than guessing behavior.

Restrictions:

- Do not make `schemaLoss` the final validator.
- Do not make it non-deterministic or dependent on map iteration order.
- Do not add broad semantic inference that can prefer invented data over preserving user-provided data.
- Do not reduce penalties without checking existing unwrap, nesting, scalar, and invalid JSON tests.
- When adding support for `oneOf`, `anyOf`, or `allOf`, score branches structurally but still require final strict validation.

## Tests

For each fixer, add public `Ensure` tests and any needed package-internal tests.

- Public tests should cover the intended end-to-end repair, a recursive case when traversal is involved, and a conservative rejection case.
- Internal tests are useful for one-change-per-invocation behavior, candidate ordering, and direct `schemaLoss` ordering.
- Include combined-pipeline tests when a fixer depends on an earlier stage changing the input first.
- Error assertions should use existing helpers such as `assertInvalidJSONError` or `assertSchemaViolationError` when available.

## Validation Commands

Run focused tests first:

```sh
go test ./schema_compliance
```

Run the full repo check before finishing:

```sh
make
```

`make` is the default full process. It runs:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go build ./...`
- demo app builds under `cmd/...`
