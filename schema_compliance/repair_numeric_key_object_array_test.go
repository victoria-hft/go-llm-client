package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const stringArraySchema = `{
  "type": "array",
  "items": {"type": "string"}
}`

func TestEnsureRepairsZeroBasedNumericKeyObjectToArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"0":"a","1":"b"}`, stringArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["a","b"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["a","b"]`)
	}
}

func TestEnsureRepairsOneBasedNumericKeyObjectToArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"1":"a","2":"b"}`, stringArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["a","b"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["a","b"]`)
	}
}

func TestEnsureRepairsNumericKeyObjectToArrayRecursively(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"tags":{"0":"a","1":"b"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["a","b"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["a","b"]}`)
	}
}

func TestEnsureRepairsNumericKeyObjectToArrayUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "array",
	      "items": {"type": "string"}
	    },
	    {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`{"0":"a","1":"b"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["a","b"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["a","b"]`)
	}
}

func TestEnsureDoesNotRepairNumericKeyObjectWithInvalidArrayKeys(t *testing.T) {
	tests := []string{
		`{"0":"a","2":"b"}`,
		`{"0":"a","name":"b"}`,
		`{"-1":"a","0":"b"}`,
		`{"01":"a","02":"b"}`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := schema_compliance.Ensure(input, stringArraySchema)
			assertSchemaViolationError(t, err)
		})
	}
}
