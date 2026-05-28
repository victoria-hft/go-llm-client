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

const sentimentEnumSchema = `{
  "type": "object",
  "required": ["sentiment"],
  "properties": {
    "sentiment": {
      "type": "string",
      "enum": ["positive", "neutral", "negative"]
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

func TestEnsureRepairsEnumStringExplanationPrefix(t *testing.T) {
	tests := map[string]string{
		"positive — customer is satisfied":  `{"sentiment":"positive"}`,
		"Positive: customer is satisfied":   `{"sentiment":"positive"}`,
		"POSITIVE. customer is satisfied":   `{"sentiment":"positive"}`,
		"positive - customer is satisfied":  `{"sentiment":"positive"}`,
		"positive -- customer is satisfied": `{"sentiment":"positive"}`,
		"positive – customer is satisfied":  `{"sentiment":"positive"}`,
		"positive; customer is satisfied":   `{"sentiment":"positive"}`,
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"sentiment":"`+input+`"}`, sentimentEnumSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsEnumStringExplanationSuffix(t *testing.T) {
	tests := map[string]string{
		"customer is satisfied — positive": `{"sentiment":"positive"}`,
		"customer is satisfied: Positive":  `{"sentiment":"positive"}`,
		"customer is satisfied. POSITIVE":  `{"sentiment":"positive"}`,
		"customer is satisfied; positive":  `{"sentiment":"positive"}`,
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"sentiment":"`+input+`"}`, sentimentEnumSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
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

func TestEnsureRepairsEnumExplanationRecursivelyInNestedObject(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["review"],
	  "properties": {
	    "review": {
	      "type": "object",
	      "required": ["sentiment"],
	      "properties": {
	        "sentiment": {
	          "type": "string",
	          "enum": ["positive", "neutral", "negative"]
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"review":{"sentiment":"Positive: customer is satisfied"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"review":{"sentiment":"positive"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"review":{"sentiment":"positive"}}`)
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

func TestEnsureRepairsEnumExplanationInArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["reviews"],
	  "properties": {
	    "reviews": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["sentiment"],
	        "properties": {
	          "sentiment": {
	            "type": "string",
	            "enum": ["positive", "neutral", "negative"]
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"reviews":[{"sentiment":"Positive: customer is satisfied"},{"sentiment":"negative — customer is unhappy"}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"reviews":[{"sentiment":"positive"},{"sentiment":"negative"}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"reviews":[{"sentiment":"positive"},{"sentiment":"negative"}]}`)
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

func TestEnsureRepairsEnumExplanationUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["sentiment"],
	      "properties": {
	        "sentiment": {
	          "type": "string",
	          "enum": ["positive", "neutral", "negative"]
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

	got, err := schema_compliance.Ensure(`{"sentiment":"customer is satisfied — Positive"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"sentiment":"positive"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"sentiment":"positive"}`)
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

func TestEnsureRepairsMultipleEnumExplanationsThroughSchemaLoop(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["primary", "secondary"],
	  "properties": {
	    "primary": {"type": "string", "enum": ["positive", "neutral", "negative"]},
	    "secondary": {"type": "string", "enum": ["positive", "neutral", "negative"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"primary":"Positive: customer is satisfied","secondary":"customer is unhappy — negative"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"primary":"positive","secondary":"negative"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"primary":"positive","secondary":"negative"}`)
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

func TestEnsureDoesNotRepairUnsafeEnumExplanationStrings(t *testing.T) {
	tests := []string{
		"the result is positive",
		"positive and negative",
		"positive-ish",
		"positive/customer",
		"positive — negative",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := schema_compliance.Ensure(`{"sentiment":"`+input+`"}`, sentimentEnumSchema)
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

func TestEnsureDoesNotRepairAmbiguousNormalizedEnumExplanation(t *testing.T) {
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

	_, err := schema_compliance.Ensure(`{"status":"In Progress: currently active"}`, schema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotChangeAlreadyValidEnumValue(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"sentiment":"positive"}`, sentimentEnumSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"sentiment":"positive"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"sentiment":"positive"}`)
	}
}
