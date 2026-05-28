package schema_compliance

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const testBasicObjectSchema = `{
  "type": "object",
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string"
    }
  },
  "additionalProperties": false
}`

func TestSchemaLossIsZeroForValidJSON(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	if loss := schemaLoss(`{"name":"Ada"}`, schema); loss != 0 {
		t.Fatalf("schemaLoss() = %d, want 0", loss)
	}
}

func TestSchemaLossScoresWrappedObjectWorseThanPayload(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	wrappedLoss := schemaLoss(`{"data":{"name":"Ada"}}`, schema)
	payloadLoss := schemaLoss(`{"name":"Ada"}`, schema)
	if wrappedLoss <= payloadLoss {
		t.Fatalf("wrapped loss = %d, payload loss = %d, want wrapped > payload", wrappedLoss, payloadLoss)
	}
}

func TestSchemaLossScoresWrappedNearMissObjectWorseThanNearMissPayload(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	wrappedLoss := schemaLoss(`{"data":{"name":42}}`, schema)
	payloadLoss := schemaLoss(`{"name":42}`, schema)
	if wrappedLoss <= payloadLoss {
		t.Fatalf("wrapped loss = %d, payload loss = %d, want wrapped > payload", wrappedLoss, payloadLoss)
	}
}

func TestSchemaLossDoesNotPreferWrappedScalarForObjectSchema(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	wrappedLoss := schemaLoss(`{"data":42}`, schema)
	payloadLoss := schemaLoss(`42`, schema)
	if wrappedLoss > payloadLoss {
		t.Fatalf("wrapped loss = %d, payload loss = %d, want wrapped <= payload", wrappedLoss, payloadLoss)
	}
}

func TestSchemaLossScoresSingleItemArrayWorseThanPayload(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	arrayLoss := schemaLoss(`[{"name":"Ada"}]`, schema)
	payloadLoss := schemaLoss(`{"name":"Ada"}`, schema)
	if arrayLoss <= payloadLoss {
		t.Fatalf("array loss = %d, payload loss = %d, want array > payload", arrayLoss, payloadLoss)
	}
}

func TestSchemaLossScoresSingleItemNearMissArrayWorseThanNearMissPayload(t *testing.T) {
	schema := mustCompileTestSchema(t, testBasicObjectSchema)

	arrayLoss := schemaLoss(`[{"name":42}]`, schema)
	payloadLoss := schemaLoss(`{"name":42}`, schema)
	if arrayLoss <= payloadLoss {
		t.Fatalf("array loss = %d, payload loss = %d, want array > payload", arrayLoss, payloadLoss)
	}
}

func TestSchemaLossScoresInvalidFormatHigherThanValidFormat(t *testing.T) {
	const schemaJSON = `{
	  "type": "object",
	  "required": ["date"],
	  "properties": {"date": {"type": "string", "format": "date"}},
	  "additionalProperties": false
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	invalidLoss := schemaLoss(`{"date":"28 May 2026"}`, schema)
	validLoss := schemaLoss(`{"date":"2026-05-28"}`, schema)
	if invalidLoss <= validLoss {
		t.Fatalf("invalid loss = %d, valid loss = %d, want invalid > valid", invalidLoss, validLoss)
	}
}

func TestSchemaLossUsesClosestOneOfBranch(t *testing.T) {
	const schemaJSON = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["date"],
	      "properties": {"date": {"type": "string", "format": "date"}},
	      "additionalProperties": false
	    },
	    {
	      "type": "object",
	      "required": ["count"],
	      "properties": {"count": {"type": "integer"}},
	      "additionalProperties": false
	    }
	  ]
	}`
	schema := mustCompileTestSchema(t, schemaJSON)

	stringLoss := schemaLoss(`{"count":"42"}`, schema)
	numberLoss := schemaLoss(`{"count":42}`, schema)
	unrelatedLoss := schemaLoss(`{"other":42}`, schema)
	if stringLoss <= numberLoss {
		t.Fatalf("string loss = %d, number loss = %d, want string > number", stringLoss, numberLoss)
	}
	if stringLoss >= unrelatedLoss {
		t.Fatalf("string loss = %d, unrelated loss = %d, want string < unrelated", stringLoss, unrelatedLoss)
	}
}

func mustCompileTestSchema(t *testing.T, schemaJSON string) *jsonschema.Schema {
	t.Helper()
	schema, err := compileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("compileSchema returned error: %v", err)
	}
	return schema
}
