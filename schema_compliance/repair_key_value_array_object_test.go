package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const integerMapSchema = `{
  "type": "object",
  "additionalProperties": {"type": "integer"}
}`

func TestEnsureRepairsKeyValueArrayToObject(t *testing.T) {
	got, err := schema_compliance.Ensure(`[{"key":"a","value":1},{"key":"b","value":2}]`, integerMapSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"a":1,"b":2}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"a":1,"b":2}`)
	}
}

func TestEnsureRepairsRepresentativeKeyValueArrayFieldNamesToObject(t *testing.T) {
	tests := map[string]string{
		"name_value":     `[{"name":"a","value":1}]`,
		"field_value":    `[{"field":"a","value":1}]`,
		"property_value": `[{"property":"a","value":1}]`,
		"id_value":       `[{"id":"a","value":1}]`,
		"key_val":        `[{"key":"a","val":1}]`,
		"name_val":       `[{"name":"a","val":1}]`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := schema_compliance.Ensure(input, integerMapSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != `{"a":1}` {
				t.Fatalf("Ensure() = %q, want %q", got, `{"a":1}`)
			}
		})
	}
}

func TestEnsureRepairsKeyValueArrayToObjectRecursively(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["scores"],
	  "properties": {
	    "scores": {
	      "type": "object",
	      "additionalProperties": {"type": "integer"}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"scores":[{"key":"a","value":1}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"scores":{"a":1}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"scores":{"a":1}}`)
	}
}

func TestEnsureRepairsKeyValueArrayToObjectUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "additionalProperties": {"type": "integer"}
	    },
	    {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`[{"key":"a","value":1}]`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"a":1}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"a":1}`)
	}
}

func TestEnsureDoesNotRepairUnsafeKeyValueArraysToObject(t *testing.T) {
	tests := map[string]string{
		"duplicate_keys":        `[{"key":"a","value":1},{"key":"a","value":2}]`,
		"non_string_key":        `[{"key":1,"value":2},{"key":3,"value":4}]`,
		"missing_key":           `[{"value":1},{"value":2}]`,
		"missing_value":         `[{"key":"a"}]`,
		"multiple_key_fields":   `[{"key":"a","name":"b","value":1}]`,
		"multiple_value_fields": `[{"key":"a","value":1,"val":2}]`,
		"extra_field":           `[{"key":"a","value":1,"other":2}]`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := schema_compliance.Ensure(input, integerMapSchema)
			assertSchemaViolationError(t, err)
		})
	}
}
