package schema_compliance_test

import (
	"errors"
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const ndjsonObjectArraySchema = `{
  "type": "array",
  "items": {
    "type": "object",
    "required": ["a"],
    "properties": {
      "a": {"type": "integer"}
    },
    "additionalProperties": false
  }
}`

const ndjsonStringArraySchema = `{
  "type": "array",
  "items": {"type": "string"}
}`

func TestEnsureRepairsNDJSONObjectLinesToArray(t *testing.T) {
	got, err := schema_compliance.Ensure("{\"a\":1}\n{\"a\":2}", ndjsonObjectArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `[{"a":1},{"a":2}]` {
		t.Fatalf("Ensure() = %q, want %q", got, `[{"a":1},{"a":2}]`)
	}
}

func TestEnsureRepairsNDJSONPrimitiveLinesToArray(t *testing.T) {
	got, err := schema_compliance.Ensure("\"a\"\n\"b\"", ndjsonStringArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["a","b"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["a","b"]`)
	}
}

func TestEnsureRepairsNDJSONUsingOneOfArrayBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    },
	    {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure("\"a\"\n\"b\"", schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["a","b"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["a","b"]`)
	}
}

func TestEnsureRepairsFencedNDJSONToArray(t *testing.T) {
	input := "Here are the rows:\n```json\n{\"a\":1}\n{\"a\":2}\n```\nDone."
	got, err := schema_compliance.Ensure(input, ndjsonObjectArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `[{"a":1},{"a":2}]` {
		t.Fatalf("Ensure() = %q, want %q", got, `[{"a":1},{"a":2}]`)
	}
}

func TestEnsureDoesNotRepairNDJSONWhenSchemaExpectsObject(t *testing.T) {
	_, err := schema_compliance.Ensure("{\"name\":\"Ada\"}\n{\"name\":\"Grace\"}", basicObjectSchema)
	assertNDJSONErrorKind(t, err, schema_compliance.ErrorKindInvalidJSON)
}

func TestEnsureDoesNotRepairNDJSONWithInvalidLine(t *testing.T) {
	_, err := schema_compliance.Ensure("{\"a\":1}\n{\"a\":", ndjsonObjectArraySchema)
	assertNDJSONErrorKind(t, err, schema_compliance.ErrorKindInvalidJSON)
}

func TestEnsureReturnsSchemaViolationForUnrepairableNDJSONItem(t *testing.T) {
	_, err := schema_compliance.Ensure("{\"a\":1}\n{\"a\":\"two\"}", ndjsonObjectArraySchema)
	assertNDJSONErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureDoesNotRepairNDJSONWithProseLine(t *testing.T) {
	_, err := schema_compliance.Ensure("{\"a\":1}\nnot json\n{\"a\":2}", ndjsonObjectArraySchema)
	assertNDJSONErrorKind(t, err, schema_compliance.ErrorKindInvalidJSON)
}

func TestEnsureDoesNotRepairSingleLineJSONObjectAsNDJSON(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"a":1}`, ndjsonStringArraySchema)
	assertNDJSONErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func assertNDJSONErrorKind(t *testing.T, err error, want schema_compliance.ErrorKind) {
	t.Helper()
	if err == nil {
		t.Fatal("Ensure returned nil error")
	}

	var complianceErr *schema_compliance.Error
	if !errors.As(err, &complianceErr) {
		t.Fatalf("error type = %T, want *schema_compliance.Error", err)
	}
	if complianceErr.Kind != want {
		t.Fatalf("error kind = %v, want %v", complianceErr.Kind, want)
	}
}
