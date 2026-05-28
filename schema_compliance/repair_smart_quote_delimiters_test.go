package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureRepairsSmartQuoteJSONDelimiters(t *testing.T) {
	got, err := schema_compliance.Ensure(`{ “name”: “Ada” }`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsSmartQuoteStringValuesOnlyAtDelimiters(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["text"],
	  "properties": {"text": {"type": "string"}},
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"text": “Ada said “hello””}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"text":"Ada said “hello”"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"text":"Ada said “hello”"}`)
	}
}

func TestEnsureDoesNotRepairSmartQuotesInNonJSONProse(t *testing.T) {
	_, err := schema_compliance.Ensure(`Here is “not json”`, basicObjectSchema)
	assertInvalidJSONError(t, err)
}
