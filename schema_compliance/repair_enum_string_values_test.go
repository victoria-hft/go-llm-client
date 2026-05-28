package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const statusEnumSchema = `{
  "type": "object",
  "required": ["status"],
  "properties": {
    "status": {
      "type": "string",
      "enum": ["ready", "in-progress", "done"]
    }
  },
  "additionalProperties": false
}`

func TestEnsureRepairsEnumStringCaseOnlyMismatch(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"status":"READY"}`, statusEnumSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"status":"ready"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"status":"ready"}`)
	}
}

func TestEnsureRepairsEnumStringSeparatorMismatch(t *testing.T) {
	tests := []string{"in_progress", "In Progress", "IN-PROGRESS"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"status":"`+input+`"}`, statusEnumSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != `{"status":"in-progress"}` {
				t.Fatalf("Ensure() = %q, want %q", got, `{"status":"in-progress"}`)
			}
		})
	}
}

func TestEnsureRepairsEnumStringRecursivelyInNestedObject(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["task"],
	  "properties": {
	    "task": {
	      "type": "object",
	      "required": ["status"],
	      "properties": {
	        "status": {
	          "type": "string",
	          "enum": ["ready", "in-progress", "done"]
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"task":{"status":"In Progress"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"task":{"status":"in-progress"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"task":{"status":"in-progress"}}`)
	}
}

func TestEnsureRepairsEnumStringInArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tasks"],
	  "properties": {
	    "tasks": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["status"],
	        "properties": {
	          "status": {
	            "type": "string",
	            "enum": ["ready", "in-progress", "done"]
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"tasks":[{"status":"READY"},{"status":"In Progress"}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tasks":[{"status":"ready"},{"status":"in-progress"}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tasks":[{"status":"ready"},{"status":"in-progress"}]}`)
	}
}

func TestEnsureRepairsEnumStringUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["status"],
	      "properties": {
	        "status": {
	          "type": "string",
	          "enum": ["in-progress", "done"]
	        }
	      },
	      "additionalProperties": false
	    },
	    {
	      "type": "object",
	      "required": ["count"],
	      "properties": {"count": {"type": "integer"}},
	      "additionalProperties": false
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`{"status":"IN_PROGRESS"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"status":"in-progress"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"status":"in-progress"}`)
	}
}

func TestEnsureRepairsMultipleEnumStringsThroughSchemaLoop(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["primary", "secondary"],
	  "properties": {
	    "primary": {"type": "string", "enum": ["ready", "in-progress", "done"]},
	    "secondary": {"type": "string", "enum": ["ready", "in-progress", "done"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"primary":"READY","secondary":"In Progress"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"primary":"ready","secondary":"in-progress"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"primary":"ready","secondary":"in-progress"}`)
	}
}

func TestEnsureDoesNotRepairNonStringEnumValue(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status"],
	  "properties": {
	    "status": {
	      "enum": ["1"]
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"status":1}`, schema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEnumStringWithExtraPunctuationOrWords(t *testing.T) {
	tests := []string{"ready!", "ready now"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := schema_compliance.Ensure(`{"status":"`+input+`"}`, statusEnumSchema)
			assertSchemaViolationError(t, err)
		})
	}
}

func TestEnsureDoesNotRepairAmbiguousNormalizedEnumString(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "in_progress"]
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"status":"In Progress"}`, schema)
	assertSchemaViolationError(t, err)
}
