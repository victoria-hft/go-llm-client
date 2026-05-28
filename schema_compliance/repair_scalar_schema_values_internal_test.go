package schema_compliance

import "testing"

func TestRepairScalarSchemaValuesMakesOneChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["start", "end"],
	  "properties": {
	    "start": {"type": "string", "format": "date"},
	    "end": {"type": "string", "format": "date"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairScalarSchemaValues(`{"start":"28 May 2026","end":"29 May 2026"}`, schema)
	if !changed {
		t.Fatal("repairScalarSchemaValues did not change input")
	}
	want := `{"end":"2026-05-29","start":"28 May 2026"}`
	if got != want {
		t.Fatalf("repairScalarSchemaValues() = %q, want %q", got, want)
	}
}

func TestRepairScalarSchemaValuesDeclinesCandidateWhenLossDoesNotImprove(t *testing.T) {
	schema := mustCompileTestSchema(t, basicObjectSchemaForInternalScalarTests)

	_, changed := repairScalarSchemaValues(`{"name":"42"}`, schema)
	if changed {
		t.Fatal("repairScalarSchemaValues changed an already-valid arbitrary string")
	}
}

func TestRepairScalarSchemaValuesDoesNotGuessAmbiguousNumericDate(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["date"],
	  "properties": {
	    "date": {"type": "string", "format": "date"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	_, changed := repairScalarSchemaValues(`{"date":"05/06/2026"}`, schema)
	if changed {
		t.Fatal("repairScalarSchemaValues repaired an ambiguous numeric date")
	}
}

func TestRepairScalarSchemaValuesMakesOneEpochChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["start", "end"],
	  "properties": {
	    "start": {"type": "string", "format": "date-time"},
	    "end": {"type": "string", "format": "date-time"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairScalarSchemaValues(`{"start":1779975900,"end":1780062300}`, schema)
	if !changed {
		t.Fatal("repairScalarSchemaValues did not change input")
	}
	want := `{"end":"2026-05-29T13:45:00Z","start":1779975900}`
	if got != want {
		t.Fatalf("repairScalarSchemaValues() = %q, want %q", got, want)
	}
}

func TestRepairScalarSchemaValuesDoesNotRepairDurationWithoutFormat(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["duration"],
	  "properties": {
	    "duration": {"type": "string"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	_, changed := repairScalarSchemaValues(`{"duration":"5 minutes"}`, schema)
	if changed {
		t.Fatal("repairScalarSchemaValues repaired duration text without duration format")
	}
}

const basicObjectSchemaForInternalScalarTests = `{
  "type": "object",
  "required": ["name"],
  "properties": {
    "name": {"type": "string"}
  },
  "additionalProperties": false
}`
