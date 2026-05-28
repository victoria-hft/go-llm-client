package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const nestedLocationSchema = `{
  "type": "object",
  "required": ["location"],
  "properties": {
    "location": {
      "type": "object",
      "required": ["city", "country"],
      "properties": {
        "city": {"type": "string"},
        "country": {"type": "string"}
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}`

const flatLocationSchema = `{
  "type": "object",
  "required": ["city", "country"],
  "properties": {
    "city": {"type": "string"},
    "country": {"type": "string"}
  },
  "additionalProperties": false
}`

func TestEnsureRepairsPromotedSubfields(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"city":"Paris","country":"France"}`, nestedLocationSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"location":{"city":"Paris","country":"France"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"location":{"city":"Paris","country":"France"}}`)
	}
}

func TestEnsureRepairsPromotedSubfieldsWithGenericNames(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["profile"],
	  "properties": {
	    "profile": {
	      "type": "object",
	      "required": ["first", "last"],
	      "properties": {
	        "first": {"type": "string"},
	        "last": {"type": "string"}
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"first":"Ada","last":"Lovelace"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"profile":{"first":"Ada","last":"Lovelace"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"profile":{"first":"Ada","last":"Lovelace"}}`)
	}
}

func TestEnsureRepairsNestedFieldsWhenSchemaExpectsParentFields(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"location":{"city":"Paris","country":"France"}}`, flatLocationSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"city":"Paris","country":"France"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"city":"Paris","country":"France"}`)
	}
}

func TestEnsureRepairsNestedFieldsWhenSchemaExpectsParentFieldsWithGenericNames(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["first", "last"],
	  "properties": {
	    "first": {"type": "string"},
	    "last": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"profile":{"first":"Ada","last":"Lovelace"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"first":"Ada","last":"Lovelace"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"first":"Ada","last":"Lovelace"}`)
	}
}

func TestEnsureRepairsObjectFieldNestingRecursively(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["event"],
	  "properties": {
	    "event": {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
	        "location": {
	          "type": "object",
	          "required": ["city", "country"],
	          "properties": {
	            "city": {"type": "string"},
	            "country": {"type": "string"}
	          },
	          "additionalProperties": false
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"event":{"city":"Paris","country":"France"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"event":{"location":{"city":"Paris","country":"France"}}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"event":{"location":{"city":"Paris","country":"France"}}}`)
	}
}

func TestEnsureRepairsObjectFieldNestingUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    },
	    {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
	        "location": {
	          "type": "object",
	          "required": ["city", "country"],
	          "properties": {
	            "city": {"type": "string"},
	            "country": {"type": "string"}
	          },
	          "additionalProperties": false
	        }
	      },
	      "additionalProperties": false
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`{"city":"Paris","country":"France"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"location":{"city":"Paris","country":"France"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"location":{"city":"Paris","country":"France"}}`)
	}
}

func TestEnsureDoesNotOverwriteExistingDestinationObjectFields(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"location":{"city":"Lyon"},"city":"Paris","country":"France"}`, nestedLocationSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureIgnoresObjectFieldNestingCandidateWhenLossDoesNotImprove(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"details":{"city":"Paris"}}`, nestedLocationSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureSchemaStageLoopHandlesOneNestingMovePerInvocation(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["contact"],
	  "properties": {
	    "contact": {
	      "type": "object",
	      "required": ["address"],
	      "properties": {
	        "address": {
	          "type": "object",
	          "required": ["city", "country"],
	          "properties": {
	            "city": {"type": "string"},
	            "country": {"type": "string"}
	          },
	          "additionalProperties": false
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"contact":{"city":"Paris","country":"France"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"contact":{"address":{"city":"Paris","country":"France"}}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"contact":{"address":{"city":"Paris","country":"France"}}}`)
	}
}
