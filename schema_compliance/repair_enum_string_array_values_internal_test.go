package schema_compliance

import "testing"

func TestRepairEnumStringArrayValuesMakesOneChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["primary", "secondary"],
	  "properties": {
	    "primary": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["bar", "baz"]}
	    },
	    "secondary": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["abc", "done"]}
	    }
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairEnumStringArrayValues(`{"primary":"bar,baz","secondary":"abc,done"}`, schema)
	if !changed {
		t.Fatal("repairEnumStringArrayValues did not change input")
	}
	want := `{"primary":["bar","baz"],"secondary":"abc,done"}`
	if got != want {
		t.Fatalf("repairEnumStringArrayValues() = %q, want %q", got, want)
	}
}

func TestRepairEnumStringArrayValuesDeclinesCandidateWhenLossDoesNotImprove(t *testing.T) {
	schema := mustCompileTestSchema(t, `{
	  "type": "object",
	  "required": ["name"],
	  "properties": {"name": {"type": "string"}},
	  "additionalProperties": false
	}`)

	_, changed := repairEnumStringArrayValues(`{"name":"bar,baz"}`, schema)
	if changed {
		t.Fatal("repairEnumStringArrayValues changed a plain string field")
	}
}
