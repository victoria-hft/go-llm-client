package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const objectItemsSchema = `{
  "type": "object",
  "required": ["items"],
  "properties": {
    "items": {
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

const stringItemsSchema = `{
  "type": "object",
  "required": ["items"],
  "properties": {
    "items": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "additionalProperties": false
}`

const objectItemSchema = `{
  "type": "object",
  "required": ["item"],
  "properties": {
    "item": {
      "type": "object",
      "required": ["name"],
      "properties": {"name": {"type": "string"}},
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}`

const nullableStringItemSchema = `{
  "type": "object",
  "required": ["item"],
  "properties": {
    "item": {"type": ["string", "null"]}
  },
  "additionalProperties": false
}`

func TestEnsureRepairsItemToItemsArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"item":{"name":"Ada"}}`, objectItemsSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":[{"name":"Ada"}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":[{"name":"Ada"}]}`)
	}
}

func TestEnsureRepairsNullItemToEmptyItemsArray(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"item":null}`, stringItemsSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":[]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":[]}`)
	}
}

func TestEnsureRepairsSingleItemsArrayToItem(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"items":[{"name":"Ada"}]}`, objectItemSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"item":{"name":"Ada"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"item":{"name":"Ada"}}`)
	}
}

func TestEnsureRepairsEmptyItemsArrayToNullableItem(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"items":[]}`, nullableStringItemSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"item":null}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"item":null}`)
	}
}

func TestEnsureRepairsNonArrayItemsValueToItem(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"items":{"name":"Ada"}}`, objectItemSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"item":{"name":"Ada"}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"item":{"name":"Ada"}}`)
	}
}

func TestEnsureRepairsItemItemsShapeRecursivelyInObject(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["group"],
	  "properties": {
	    "group": {
	      "type": "object",
	      "required": ["items"],
	      "properties": {
	        "items": {
	          "type": "array",
	          "items": {"type": "string"}
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"group":{"item":"Ada"}}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"group":{"items":["Ada"]}}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"group":{"items":["Ada"]}}`)
	}
}

func TestEnsureRepairsItemItemsShapeRecursivelyInArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["groups"],
	  "properties": {
	    "groups": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["items"],
	        "properties": {
	          "items": {
	            "type": "array",
	            "items": {"type": "string"}
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"groups":[{"item":"Ada"}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"groups":[{"items":["Ada"]}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"groups":[{"items":["Ada"]}]}`)
	}
}

func TestEnsureRepairsItemItemsShapeUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["items"],
	      "properties": {
	        "items": {
	          "type": "array",
	          "items": {"type": "string"}
	        }
	      },
	      "additionalProperties": false
	    },
	    {
	      "type": "object",
	      "required": ["name"],
	      "properties": {"name": {"type": "string"}},
	      "additionalProperties": false
	    }
	  ]
	}`

	got, err := schema_compliance.Ensure(`{"item":"Ada"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"items":["Ada"]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"items":["Ada"]}`)
	}
}

func TestEnsureDoesNotRepairItemItemsShapeOverExistingDestination(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"item":"Ada","items":[]}`, stringItemsSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairEmptyItemsArrayWhenItemDoesNotAllowNull(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"items":[]}`, objectItemSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairMultiItemArrayToSingularItem(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"items":[{"name":"Ada"},{"name":"Grace"}]}`, objectItemSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureDoesNotRepairItemItemsShapeWhenLossDoesNotImprove(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"item":"Ada"}`, basicObjectSchema)
	assertSchemaViolationError(t, err)
}
