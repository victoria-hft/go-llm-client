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

const relaxedStringArraySchema = `{
  "type": "array",
  "items": {"type": "string"}
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

func TestEnsureRepairsRelaxedJSONBareIdentifierArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`[risk, pricing]`, relaxedStringArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["risk","pricing"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["risk","pricing"]`)
	}
}

func TestEnsureRepairsRelaxedJSONBareIdentifierArrayProperty(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{tags:[risk, pricing]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"tags":["risk","pricing"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"tags":["risk","pricing"]}`)
	}
}

func TestEnsureRepairsRelaxedJSONNestedBareIdentifierArrays(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["groups"],
	  "properties": {
	    "groups": {
	      "type": "array",
	      "items": {
	        "type": "array",
	        "items": {"type": "string"}
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{groups:[[risk, pricing]]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"groups":[["risk","pricing"]]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"groups":[["risk","pricing"]]}`)
	}
}

func TestEnsureRepairsRelaxedJSONMixedBareAndQuotedStringArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`[risk, "pricing"]`, relaxedStringArraySchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `["risk","pricing"]` {
		t.Fatalf("Ensure() = %q, want %q", got, `["risk","pricing"]`)
	}
}

func TestEnsureRepairsRelaxedJSONSpecialBareArrayValues(t *testing.T) {
	const schema = `{
	  "type": "array",
	  "items": {"type": ["boolean", "null"]}
	}`

	got, err := schema_compliance.Ensure(`[true, false, null, undefined]`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `[true,false,null,null]` {
		t.Fatalf("Ensure() = %q, want %q", got, `[true,false,null,null]`)
	}
}

func TestEnsureDoesNotRepairRelaxedJSONArrayIdentifierWithSpaces(t *testing.T) {
	_, err := schema_compliance.Ensure(`[high risk, pricing]`, relaxedStringArraySchema)
	assertInvalidJSONError(t, err)
}

func TestEnsureRepairsRelaxedJSONUndefinedValueToNull(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["x"],
	  "properties": {
	    "x": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"x": undefined}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":null}`)
	}
}

func TestEnsureRepairsRelaxedJSONUndefinedValueWithUnquotedKey(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["x"],
	  "properties": {
	    "x": {"type": ["string", "null"]}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{x: undefined}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":null}`)
	}
}

func TestEnsureRepairsRelaxedJSONUndefinedNestedAndArrayValues(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items", "profile"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["value"],
	        "properties": {
	          "value": {"type": ["string", "null"]}
	        },
	        "additionalProperties": false
	      }
	    },
	    "profile": {
	      "type": "object",
	      "required": ["value"],
	      "properties": {
	        "value": {"type": ["string", "null"]}
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{items:[{value: undefined}], profile:{value: undefined}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	want := `{"items":[{"value":null}],"profile":{"value":null}}`
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func TestEnsureRepairsRelaxedJSONBareNaNValueToString(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["x"],
	  "properties": {
	    "x": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"x": NaN}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":"NaN"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":"NaN"}`)
	}
}

func TestEnsureRepairsRelaxedJSONBareNaNValueWithUnquotedKey(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["x"],
	  "properties": {
	    "x": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{x: NaN}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":"NaN"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":"NaN"}`)
	}
}

func TestEnsureKeepsNaNObjectKey(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["NaN"],
	  "properties": {
	    "NaN": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{NaN: "value"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"NaN":"value"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"NaN":"value"}`)
	}
}

func TestEnsureKeepsQuotedNaNString(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["x"],
	  "properties": {
	    "x": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"x":"NaN"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"x":"NaN"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"x":"NaN"}`)
	}
}

func TestEnsureKeepsQuotedUndefinedString(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"undefined"}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"undefined"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"undefined"}`)
	}
}

func TestEnsureKeepsUndefinedObjectKey(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["undefined"],
	  "properties": {
	    "undefined": {"type": "string"}
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{undefined: "value"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"undefined":"value"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"undefined":"value"}`)
	}
}

func TestEnsureReturnsSchemaViolationForUndefinedWhenNullNotAllowed(t *testing.T) {
	_, err := schema_compliance.Ensure(`{name: undefined}`, basicObjectSchema)
	assertSchemaViolationError(t, err)
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
