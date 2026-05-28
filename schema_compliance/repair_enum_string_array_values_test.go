package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const tagEnumArraySchema = `{
  "type": "object",
  "required": ["tags"],
  "properties": {
    "tags": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["bar", "baz", "abc", "in-progress", "done"]
      }
    }
  },
  "additionalProperties": false
}`

func TestEnsureRepairsSingleEnumStringToEnumArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"tags":"bar"}`, tagEnumArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["bar"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["bar"]}`)
	}
}

func TestEnsureRepairsCommaSeparatedEnumStringToEnumArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"tags":"bar,baz,abc"}`, tagEnumArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["bar","baz","abc"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["bar","baz","abc"]}`)
	}
}

func TestEnsureRepairsCommaSeparatedEnumStringWithWhitespace(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"tags":"bar, baz, abc"}`, tagEnumArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["bar","baz","abc"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["bar","baz","abc"]}`)
	}
}

func TestEnsureRepairsCommaSeparatedEnumStringWithTokenNormalization(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"tags":"IN_PROGRESS, Done"}`, tagEnumArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["in-progress","done"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["in-progress","done"]}`)
	}
}

func TestEnsureRepairsEnumStringArrayRecursivelyInNestedObject(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["task"],
	  "properties": {
	    "task": {
	      "type": "object",
	      "required": ["tags"],
	      "properties": {
	        "tags": {
	          "type": "array",
	          "items": {"type": "string", "enum": ["bar", "baz", "abc"]}
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"task":{"tags":"bar,baz"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"task":{"tags":["bar","baz"]}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"task":{"tags":["bar","baz"]}}`)
	}
}

func TestEnsureRepairsEnumStringArrayInArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tasks"],
	  "properties": {
	    "tasks": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["tags"],
	        "properties": {
	          "tags": {
	            "type": "array",
	            "items": {"type": "string", "enum": ["bar", "baz", "abc"]}
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"tasks":[{"tags":"bar,baz"}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tasks":[{"tags":["bar","baz"]}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tasks":[{"tags":["bar","baz"]}]}`)
	}
}

func TestEnsureRepairsEnumStringArrayUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["tags"],
	      "properties": {
	        "tags": {
	          "type": "array",
	          "items": {"type": "string", "enum": ["bar", "baz"]}
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

	got, err := schema_compliance.Ensure(`{"tags":"bar,baz"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["bar","baz"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["bar","baz"]}`)
	}
}

func TestEnsureRepairsMultipleEnumStringArraysThroughSchemaLoop(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["primary", "secondary"],
	  "properties": {
	    "primary": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["bar", "baz"]}
	    },
	    "secondary": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["abc", "done"]}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"primary":"bar,baz","secondary":"abc,done"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"primary":["bar","baz"],"secondary":["abc","done"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"primary":["bar","baz"],"secondary":["abc","done"]}`)
	}
}

func TestEnsureDoesNotRepairEnumStringArrayWithUnknownToken(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"tags":"bar,unknown"}`, tagEnumArraySchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEnumStringArrayWithEmptyToken(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"tags":"bar,,baz"}`, tagEnumArraySchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEnumStringArrayWithAmbiguousNormalizedEnum(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["in-progress", "in_progress"]}
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"tags":"In Progress"}`, schema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEnumStringArrayWithNonStringEnumMember(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["bar", null]}
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"tags":"bar"}`, schema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEnumStringArrayWhenSchemaExpectsPlainString(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"status":"bar,baz"}`, `{
	  "type": "object",
	  "required": ["status"],
	  "properties": {"status": {"type": "string"}},
	  "additionalProperties": false
	}`)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"status":"bar,baz"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"status":"bar,baz"}`)
	}
}

func TestEnsureDoesNotRepairStringArrayWithoutEnum(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"tags":"bar,baz"}`, `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`)
	assertSchemaViolationError(t, err)
}
