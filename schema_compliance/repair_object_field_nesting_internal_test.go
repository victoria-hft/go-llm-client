package schema_compliance

import "testing"

func TestRepairObjectFieldNestingMakesOneMovePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["city", "country", "first", "last"],
	  "properties": {
	    "city": {"type": "string"},
	    "country": {"type": "string"},
	    "first": {"type": "string"},
	    "last": {"type": "string"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairObjectFieldNesting(`{"location":{"city":"Paris","country":"France"},"profile":{"first":"Ada","last":"Lovelace"}}`, schema)
	if !changed {
		t.Fatal("repairObjectFieldNesting did not change input")
	}
	want := `{"city":"Paris","country":"France","profile":{"first":"Ada","last":"Lovelace"}}`
	if got != want {
		t.Fatalf("repairObjectFieldNesting() = %q, want %q", got, want)
	}
}

func TestRepairObjectFieldNestingDeclinesOverwriteCandidate(t *testing.T) {
	schema := mustCompileTestSchema(t, nestedLocationSchemaForInternalTests)

	_, changed := repairObjectFieldNesting(`{"location":{"city":"Lyon"},"city":"Paris","country":"France"}`, schema)
	if changed {
		t.Fatal("repairObjectFieldNesting changed input with an existing destination object")
	}
}

const nestedLocationSchemaForInternalTests = `{
  "type": "object",
  "required": ["location"],
  "properties": {
    "location": {
      "type": "object",
      "required": ["city", "country"],
      "properties": {
        "city": {"type": "string"},
        "country": {"type": "string"}
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}`
