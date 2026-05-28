package schema_compliance

import "testing"

func TestRepairItemItemsShapeMakesOneChangePerInvocation(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["first", "second"],
	  "properties": {
	    "first": {
	      "type": "object",
	      "required": ["items"],
	      "properties": {
	        "items": {
	          "type": "array",
	          "items": {"type": "string"}
	        }
	      },
	      "additionalProperties": false
	    },
	    "second": {
	      "type": "object",
	      "required": ["items"],
	      "properties": {
	        "items": {
	          "type": "array",
	          "items": {"type": "string"}
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	got, changed := repairItemItemsShape(`{"first":{"item":"Ada"},"second":{"item":"Grace"}}`, schema)
	if !changed {
		t.Fatal("repairItemItemsShape did not change input")
	}
	want := `{"first":{"items":["Ada"]},"second":{"item":"Grace"}}`
	if got != want {
		t.Fatalf("repairItemItemsShape() = %q, want %q", got, want)
	}
}

func TestRepairItemItemsShapeDeclinesCandidateWhenLossDoesNotImprove(t *testing.T) {
	schema := mustCompileTestSchema(t, basicObjectSchemaForInternalScalarTests)

	_, changed := repairItemItemsShape(`{"item":"Ada"}`, schema)
	if changed {
		t.Fatal("repairItemItemsShape changed input without a schema-backed improvement")
	}
}
