package schema_compliance_test

import (
	"errors"
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const basicObjectSchema = `{
  "type": "object",
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string"
    }
  },
  "additionalProperties": false
}`

func TestEnsureExtractsJSONFencedBlockWithLanguage(t *testing.T) {
	got, err := schema_compliance.Ensure("Here is the result:\n```json {\"name\":\"Ada\"} ```", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureExtractsJSONFencedBlockWithoutLanguage(t *testing.T) {
	got, err := schema_compliance.Ensure("Here is the result:\n``` {\"name\":\"Ada\"} ```", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureExtractsFencedBlockFromArbitraryShortText(t *testing.T) {
	got, err := schema_compliance.Ensure("I checked the request and this should match:\n```json\n{\"name\":\"Ada\"}\n```\nHope that helps.", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureReturnsInvalidSchemaError(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":"Ada"}`, `{"type":`)
	if err == nil {
		t.Fatal("Ensure returned nil error")
	}

	var complianceErr *schema_compliance.Error
	if !errors.As(err, &complianceErr) {
		t.Fatalf("error type = %T, want *schema_compliance.Error", err)
	}
	if complianceErr.Kind != schema_compliance.ErrorKindInvalidSchema {
		t.Fatalf("error kind = %v, want %v", complianceErr.Kind, schema_compliance.ErrorKindInvalidSchema)
	}
	if complianceErr.Unwrap() == nil {
		t.Fatal("expected wrapped schema error")
	}
}

func TestEnsureReturnsInvalidJSONError(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":`, basicObjectSchema)
	if err == nil {
		t.Fatal("Ensure returned nil error")
	}

	var complianceErr *schema_compliance.Error
	if !errors.As(err, &complianceErr) {
		t.Fatalf("error type = %T, want *schema_compliance.Error", err)
	}
	if complianceErr.Kind != schema_compliance.ErrorKindInvalidJSON {
		t.Fatalf("error kind = %v, want %v", complianceErr.Kind, schema_compliance.ErrorKindInvalidJSON)
	}
	if complianceErr.Unwrap() == nil {
		t.Fatal("expected wrapped parse error")
	}
}

func TestEnsureReturnsSchemaViolationError(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"name":42}`, basicObjectSchema)
	if err == nil {
		t.Fatal("Ensure returned nil error")
	}

	var complianceErr *schema_compliance.Error
	if !errors.As(err, &complianceErr) {
		t.Fatalf("error type = %T, want *schema_compliance.Error", err)
	}
	if complianceErr.Kind != schema_compliance.ErrorKindSchemaViolation {
		t.Fatalf("error kind = %v, want %v", complianceErr.Kind, schema_compliance.ErrorKindSchemaViolation)
	}
	if complianceErr.Unwrap() == nil {
		t.Fatal("expected wrapped validation error")
	}
}
