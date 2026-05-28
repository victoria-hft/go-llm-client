package schema_compliance_test

import (
	"errors"
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const nestedObjectSchema = `{
  "type": "object",
  "required": ["name", "tags", "profile"],
  "properties": {
    "name": {"type": "string"},
    "tags": {
      "type": "array",
      "items": {"type": "string"}
    },
    "profile": {
      "type": "object",
      "required": ["active"],
      "properties": {
        "active": {"type": "boolean"}
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}`

func TestEnsureRepairsRelaxedJSONUnquotedPropertyKey(t *testing.T) {
	got, err := schema_compliance.Ensure(`{name: "Ada"}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsRelaxedJSONSingleQuotedStrings(t *testing.T) {
	got, err := schema_compliance.Ensure(`{'name':'Ada'}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsRelaxedJSONTrailingComma(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"Ada",}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsRelaxedJSONBareIdentifierValue(t *testing.T) {
	got, err := schema_compliance.Ensure(`{name: Ada}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureRepairsRelaxedJSONNestedValuesAndTrailingCommas(t *testing.T) {
	got, err := schema_compliance.Ensure(`{name: 'Ada', tags: ['math', research,], profile: {active: true,},}`, nestedObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"name":"Ada","profile":{"active":true},"tags":["math","research"]}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureCombinesFencedAndRelaxedJSONRepairWithLanguage(t *testing.T) {
	got, err := schema_compliance.Ensure("Here is the result:\n```json {name: 'Ada'} ```", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureCombinesFencedAndRelaxedJSONRepairWithArbitraryText(t *testing.T) {
	got, err := schema_compliance.Ensure("I checked it:\n```json\n{name: 'Ada'}\n```\nDone.", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureCombinesPlainFencedAndRelaxedJSONRepair(t *testing.T) {
	got, err := schema_compliance.Ensure("```\n{name: 'Ada'}\n```", basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"Ada"}`)
	}
}

func TestEnsureDoesNotRepairRelaxedJSONWithComments(t *testing.T) {
	_, err := schema_compliance.Ensure("{name: 'Ada' // comment\n}", basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairYAMLDocument(t *testing.T) {
	_, err := schema_compliance.Ensure("---\nname: Ada", basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairPartialParse(t *testing.T) {
	_, err := schema_compliance.Ensure("{name: 'Ada'} trailing", basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureDoesNotRepairBareValueWithSpaces(t *testing.T) {
	_, err := schema_compliance.Ensure("{name: Ada Lovelace}", basicObjectSchema)
	assertInvalidJSONError(t, err)
}

func assertInvalidJSONError(t *testing.T, err error) {
	t.Helper()
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
}
