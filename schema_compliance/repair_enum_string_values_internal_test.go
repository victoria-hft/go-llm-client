package schema_compliance

import "testing"

func TestRepairEnumStringValuesMakesOneChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["primary", "secondary"],
	  "properties": {
	    "primary": {"type": "string", "enum": ["ready", "in-progress", "done"]},
	    "secondary": {"type": "string", "enum": ["ready", "in-progress", "done"]}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairEnumStringValues(`{"primary":"READY","secondary":"In Progress"}`, schema)
	if !changed {
		t.Fatal("repairEnumStringValues did not change input")
	}
	want := `{"primary":"ready","secondary":"In Progress"}`
	if got != want {
		t.Fatalf("repairEnumStringValues() = %q, want %q", got, want)
	}
}

func TestRepairEnumStringValuesDeclinesAmbiguousNormalizedEnum(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["status"],
	  "properties": {
	    "status": {"type": "string", "enum": ["in-progress", "in_progress"]}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	_, changed := repairEnumStringValues(`{"status":"In Progress"}`, schema)
	if changed {
		t.Fatal("repairEnumStringValues repaired an ambiguous enum value")
	}
}

func TestRepairEnumStringValuesDeclinesCandidateWhenLossDoesNotImprove(t *testing.T) {
	schema := mustCompileTestSchema(t, basicObjectSchemaForInternalScalarTests)

	_, changed := repairEnumStringValues(`{"name":"Ada"}`, schema)
	if changed {
		t.Fatal("repairEnumStringValues changed an already-valid arbitrary string")
	}
}
