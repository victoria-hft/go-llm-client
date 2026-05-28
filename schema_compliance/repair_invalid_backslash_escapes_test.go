package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureRejectsInvalidBackslashEscapes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["text"],
	  "properties": {"text": {"type": "string"}},
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"text":"\qword"}`, schema)
	assertInvalidJSONError(t, err)
}

func TestEnsurePreservesValidBackslashEscapes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["text"],
	  "properties": {"text": {"type": "string"}},
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"text":"line\nquote\"slash\\unicode\u03b1"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"text":"line\nquote\"slash\\unicode\u03b1"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"text":"line\nquote\"slash\\unicode\u03b1"}`)
	}
}

func TestEnsureDoesNotRepairInvalidUnicodeEscape(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["text"],
	  "properties": {"text": {"type": "string"}},
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"text":"\u12G4"}`, schema)
	assertInvalidJSONError(t, err)
}
