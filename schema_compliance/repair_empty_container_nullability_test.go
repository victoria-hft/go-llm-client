package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureRepairsEmptyObjectToNullForNullableNonObjectField(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"value":{}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":null}`)
	}
}

func TestEnsureRepairsEmptyArrayToNullForNullableNonArrayField(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["string", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"value":[]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":null}`)
	}
}

func TestEnsureRepairsNullToEmptyArrayForArrayField(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"items":null}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":[]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":[]}`)
	}
}

func TestEnsureRepairsEmptyContainersRecursively(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["profile", "rows"],
	  "properties": {
	    "profile": {
	      "type": "object",
	      "required": ["score"],
	      "properties": {
	        "score": {"type": ["number", "null"]}
	      },
	      "additionalProperties": false
	    },
	    "rows": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["tags"],
	        "properties": {
	          "tags": {
	            "type": "array",
	            "items": {"type": "string"}
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"profile":{"score":{}},"rows":[{"tags":null}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"profile":{"score":null},"rows":[{"tags":[]}]}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureRepairsEmptyContainerNullabilityThroughOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
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
	    {
	      "type": "object",
	      "required": ["value"],
	      "properties": {
	        "value": {"type": ["string", "null"]}
	      },
	      "additionalProperties": false
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`{"items":null}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":[]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":[]}`)
	}
}

func TestEnsureRepairsMultipleEmptyContainerNullabilityValues(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["first", "second", "items"],
	  "properties": {
	    "first": {"type": ["number", "null"]},
	    "second": {"type": ["string", "null"]},
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"first":{},"second":[],"items":null}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"first":null,"items":[],"second":null}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureDoesNotRepairEmptyObjectWhenObjectIsAllowed(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["object", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"value":{}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":{}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":{}}`)
	}
}

func TestEnsureDoesNotRepairEmptyArrayWhenArrayIsAllowed(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["array", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"value":[]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":[]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":[]}`)
	}
}

func TestEnsureDoesNotRepairNullWhenNullIsAllowed(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items"],
	  "properties": {
	    "items": {"type": ["array", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"items":null}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":null}`)
	}
}

func TestEnsureDoesNotRepairNonEmptyContainersToNull(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["value"],
	  "properties": {
	    "value": {"type": ["string", "null"]}
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"value":{"kept":true}}`, schema)
	assertSchemaViolationError(t, err)

	_, err = schema_compliance.Ensure(`{"value":["kept"]}`, schema)
	assertSchemaViolationError(t, err)
}
