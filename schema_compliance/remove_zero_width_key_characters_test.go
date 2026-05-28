package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureRemovesZeroWidthCharactersFromTopLevelKeys(t *testing.T) {
	got, err := schema_compliance.Ensure("{\"na\u200bme\":\"Ada\"}", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRemovesZeroWidthCharactersFromNestedKeys(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["profile"],
	  "properties": {
	    "profile": {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure("{\"profile\":{\"na\u200cme\":\"Ada\"}}", schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"profile":{"name":"Ada"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"profile":{"name":"Ada"}}`)
	}
}

func TestEnsureRemovesZeroWidthCharactersFromArrayItemKeys(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["people"],
	  "properties": {
	    "people": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["name"],
	        "properties": {"name": {"type": "string"}},
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure("{\"people\":[{\"na\u200dme\":\"Ada\"}]}", schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"people":[{"name":"Ada"}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"people":[{"name":"Ada"}]}`)
	}
}

func TestEnsureRemovesMultipleZeroWidthCharactersFromKey(t *testing.T) {
	got, err := schema_compliance.Ensure("{\"n\u2060a\ufeffm\u200be\":\"Ada\"}", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureDoesNotRemoveZeroWidthCharactersFromStringValues(t *testing.T) {
	got, err := schema_compliance.Ensure("{\"name\":\"A\u200bda\"}", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != "{\"name\":\"A\u200bda\"}" {
		t.Fatalf("Ensure() = %q, want %q", got, "{\"name\":\"A\u200bda\"}")
	}
}

func TestEnsureDoesNotRemoveZeroWidthCharactersWhenKeyCollisionWouldOccur(t *testing.T) {
	_, err := schema_compliance.Ensure("{\"name\":\"Ada\",\"na\u200bme\":\"Grace\"}", basicObjectSchema)
	assertSchemaViolationError(t, err)
}
