package schema_compliance_test

import (
	"errors"
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

func TestEnsureUnwrapsCommonResponseObjectKeys(t *testing.T) {
	tests := []string{
		"data",
		"result",
		"response",
		"output",
		"answer",
		"payload",
		"items",
	}

	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"`+key+`":{"name":"Ada"}}`, basicObjectSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != `{"name":"Ada"}` {
				t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
			}
		})
	}
}

func TestEnsureUnwrapsSingleItemArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`[{"name":"Ada"}]`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureDoesNotUnwrapAlreadyCompliantJSON(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["data"],
	  "properties": {
	    "data": {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"data":{"name":"Ada"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"data":{"name":"Ada"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"data":{"name":"Ada"}}`)
	}
}

func TestEnsureUnwrapsResponseObjectEvenWhenPayloadStillViolatesSchema(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"data":{"name":42}}`, basicObjectSchema)
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
}

func TestEnsureDoesNotUnwrapMultiKeyResponseObject(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"data":{"name":"Ada"},"meta":{"request_id":"1"}}`, basicObjectSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotUnwrapMultiItemArray(t *testing.T) {
	_, err := schema_compliance.Ensure(`[{"name":"Ada"},{"name":"Grace"}]`, basicObjectSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotUnwrapWhenLossDoesNotImprove(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"data":42}`, basicObjectSchema)
	assertSchemaViolationError(t, err)
}

func assertSchemaViolationError(t *testing.T, err error) {
	t.Helper()
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
}
