package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureRepairsTruncatedObjectMissingFinalBrace(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"Ada"`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsTruncatedNestedObjectMissingFinalBraces(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"Ada","tags":["math"],"profile":{"active":true`, nestedObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"name":"Ada","profile":{"active":true},"tags":["math"]}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureRepairsTruncatedArrayMissingFinalBracket(t *testing.T) {
	const schema = `{
	  "type": "array",
	  "items": {"type": "string"}
	}`

	got, err := schema_compliance.Ensure(`["math","research"`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["math","research"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["math","research"]`)
	}
}

func TestEnsureRepairsTruncatedObjectAfterNestedArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"Ada","tags":["math","research"],"profile":{"active":true}`, nestedObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"name":"Ada","profile":{"active":true},"tags":["math","research"]}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureDoesNotRepairTruncatedJSONInsideString(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":"Ada`, basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairTruncatedJSONMissingValue(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":`, basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairTruncatedJSONMissingColon(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name"`, basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairTruncatedJSONAfterComma(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":"Ada",`, basicObjectSchema)
	assertInvalidJSONError(t, err)
}
