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

func TestRepairScalarSchemaValuesMakesOneNumericChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["first", "second"],
	  "properties": {
	    "first": {"type": "number"},
	    "second": {"type": "number"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairScalarSchemaValues(`{"first":"1_000","second":"3/4"}`, schema)
	if !changed {
		t.Fatal("repairScalarSchemaValues did not change input")
	}
	want := `{"first":1000,"second":"3/4"}`
	if got != want {
		t.Fatalf("repairScalarSchemaValues() = %q, want %q", got, want)
	}
}

func TestRepairScalarSchemaValuesDoesNotRepairPercentWithoutBounds(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": "number"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	_, changed := repairScalarSchemaValues(`{"value":"92%"}`, schema)
	if changed {
		t.Fatal("repairScalarSchemaValues repaired percent without schema bounds")
	}
}

func TestRepairScalarSchemaValuesMakesOneNaNChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["first", "second"],
	  "properties": {
	    "first": {"type": ["number", "null"]},
	    "second": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairScalarSchemaValues(`{"first":"NaN","second":"NaN"}`, schema)
	if !changed {
		t.Fatal("repairScalarSchemaValues did not repair nullable numeric NaN")
	}
	want := `{"first":null,"second":"NaN"}`
	if got != want {
		t.Fatalf("repairScalarSchemaValues() = %q, want %q", got, want)
	}
}

func TestRepairScalarSchemaValuesMakesOneUUIDChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["first", "second"],
	  "properties": {
	    "first": {"type": "string"},
	    "second": {"type": "string"}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairScalarSchemaValues(`{"first":"F0307D30EAD2417392873DAB7FFA0FA4","second":"F0307D30EAD2417392873DAB7FFA0FA4"}`, schema)
	if !changed {
		t.Fatal("repairScalarSchemaValues did not repair UUID string")
	}
	want := `{"first":"f0307d30-ead2-4173-9287-3dab7ffa0fa4","second":"F0307D30EAD2417392873DAB7FFA0FA4"}`
	if got != want {
		t.Fatalf("repairScalarSchemaValues() = %q, want %q", got, want)
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
