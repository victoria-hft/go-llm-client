package schema_compliance

import "testing"

func TestRepairEmptyContainerNullabilityMakesOneChangePerInvocation(t *testing.T) {
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

	got, changed := repairEmptyContainerNullability(`{"first":{},"second":{}}`, schema)
	if !changed {
		t.Fatal("repairEmptyContainerNullability did not change input")
	}
	want := `{"first":null,"second":{}}`
	if got != want {
		t.Fatalf("repairEmptyContainerNullability() = %q, want %q", got, want)
	}
}

func TestRepairEmptyContainerNullabilityDeclinesCandidateWhenLossDoesNotImprove(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["object", "null"]}
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	_, changed := repairEmptyContainerNullability(`{"value":{}}`, schema)
	if changed {
		t.Fatal("repairEmptyContainerNullability changed an already-valid empty object")
	}
}
