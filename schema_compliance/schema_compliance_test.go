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

const xNumberSchema = `{
  "type": "object",
  "required": ["x"],
  "properties": {
    "x": {
      "type": "number"
    }
  },
  "additionalProperties": false
}`

func TestEnsureStripsUTF8BOMBeforeJSON(t *testing.T) {
	got, err := schema_compliance.Ensure("\ufeff{\"x\":1}", xNumberSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":1}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":1}`)
	}
}

func TestEnsureStripsCommonTransportJunkBeforeJSON(t *testing.T) {
	tests := map[string]string{
		"mojibake_bom":       "ï»¿{\"x\":1}",
		"leading_nul":        "\x00{\"x\":1}",
		"replacement_prefix": "���{\"x\":1}",
		"mixed_junk":         "\x00\ufeff���{\"x\":1}",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := schema_compliance.Ensure(input, xNumberSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != `{"x":1}` {
				t.Fatalf("Ensure() = %q, want %q", got, `{"x":1}`)
			}
		})
	}
}

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

func TestEnsureReturnsAlreadyCompliantJSONWithoutApplyingFixes(t *testing.T) {
	const alreadyCompliant = "{\"name\":\"```json\"}"

	got, err := schema_compliance.Ensure(alreadyCompliant, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != alreadyCompliant {
		t.Fatalf("Ensure() = %q, want %q", got, alreadyCompliant)
	}
}

func TestValidateJSONAcceptsValidJSON(t *testing.T) {
	if err := schema_compliance.ValidateJSON(`{"name":"Ada"}`); err != nil {
		t.Fatalf("ValidateJSON returned error: %v", err)
	}
}

func TestValidateJSONRejectsInvalidJSON(t *testing.T) {
	if err := schema_compliance.ValidateJSON(`{"name":`); err == nil {
		t.Fatal("ValidateJSON returned nil error")
	}
}

func TestValidateAgainstSchemaAcceptsCompliantJSON(t *testing.T) {
	if err := schema_compliance.ValidateAgainstSchema(`{"name":"Ada"}`, basicObjectSchema); err != nil {
		t.Fatalf("ValidateAgainstSchema returned error: %v", err)
	}
}

func TestValidateAgainstSchemaRejectsNonCompliantJSON(t *testing.T) {
	if err := schema_compliance.ValidateAgainstSchema(`{"name":42}`, basicObjectSchema); err == nil {
		t.Fatal("ValidateAgainstSchema returned nil error")
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
