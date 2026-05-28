package schema_compliance_test

import (
	"errors"
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const pipelineProfileSchema = `{
  "type": "object",
  "required": ["name", "event", "score", "status", "tags"],
  "properties": {
    "name": {"type": "string"},
    "event": {
      "type": "object",
      "required": ["date", "location"],
      "properties": {
        "date": {"type": "string", "format": "date"},
        "location": {
          "type": "object",
          "required": ["city", "country"],
          "properties": {
            "city": {"type": "string"},
            "country": {"type": "string"}
          },
          "additionalProperties": false
        }
      },
      "additionalProperties": false
    },
    "score": {"type": "number"},
    "status": {
      "type": ["string", "null"],
      "enum": ["ready", "done", null]
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "additionalProperties": false
}`

const pipelineFlatEventSchema = `{
  "type": "object",
  "required": ["city", "country", "date", "count", "status"],
  "properties": {
    "city": {"type": "string"},
    "country": {"type": "string"},
    "date": {"type": "string", "format": "date"},
    "count": {"type": "integer"},
    "status": {
      "type": ["string", "null"],
      "enum": ["ready", "done", null]
    }
  },
  "additionalProperties": false
}`

const pipelinePeopleSchema = `{
  "type": "object",
  "required": ["people"],
  "properties": {
    "people": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "event", "score", "status"],
        "properties": {
          "name": {"type": "string"},
          "event": {
            "type": "object",
            "required": ["date", "location"],
            "properties": {
              "date": {"type": "string", "format": "date"},
              "location": {
                "type": "object",
                "required": ["city", "country"],
                "properties": {
                  "city": {"type": "string"},
                  "country": {"type": "string"}
                },
                "additionalProperties": false
              }
            },
            "additionalProperties": false
          },
          "score": {"type": "number"},
          "status": {
            "type": ["string", "null"],
            "enum": ["ready", "done", null]
          }
        },
        "additionalProperties": false
      }
    }
  },
  "additionalProperties": false
}`

const pipelineOneOfCountOrDateSchema = `{
  "oneOf": [
    {
      "type": "object",
      "required": ["count"],
      "properties": {"count": {"type": "integer"}},
      "additionalProperties": false
    },
    {
      "type": "object",
      "required": ["date"],
      "properties": {"date": {"type": "string", "format": "date"}},
      "additionalProperties": false
    }
  ]
}`

func TestEnsureFullPipelineRepairsTransportJunkFencedRelaxedWrappedNestedScalarOutput(t *testing.T) {
	const zwsp = "\u200b"

	input := "\ufeffHere is the result:\n```json\n" +
		`{
		  data: {
		    "na` + zwsp + `me": 'Ada',
		    event: {
		      date: '28 May 2026',
		      city: 'Paris',
		      country: 'France',
		    },
		    score: '42.5',
		    status: 'N/A',
		    tags: ['research',],
		  },
		}` +
		"\n```\nDone."
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsPlainFencedRelaxedScalarOutput(t *testing.T) {
	input := "Here is the value:\n```\n{count: '42'}\n```"
	want := `{"count":42}`

	assertEnsurePipeline(t, input, pipelineOneOfCountOrDateSchema, want)
}

func TestEnsureFullPipelineRepairsMojibakeTransportJunkPayloadWrapperOutput(t *testing.T) {
	input := "ï»¿```json\n{payload:{name:'Ada'}}\n```"
	want := `{"name":"Ada"}`

	assertEnsurePipeline(t, input, basicObjectSchema, want)
}

func TestEnsureFullPipelineRepairsTruncatedRelaxedZeroWidthWrappedSchemaLoop(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"data":{"na` + zwsp + `me":'Ada',event:{date:'May 28, 2026',city:'Paris',country:'France'},score:"42.5",status:"unknown",tags:['research']`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsTruncatedArrayWrapperThenSchemaOutput(t *testing.T) {
	input := `[{name:'Ada',event:{date:'28 May 2026',city:'Paris',country:'France'},score:'42.5',status:'nil',tags:['research']`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsSingleItemArrayWrapperWithRelaxedNestedScalarOutput(t *testing.T) {
	input := `[{name:'Ada',event:{date:'2026/05/28',city:'Paris',country:'France'},score:'42',status:'null',tags:['math',]}]`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42,"status":null,"tags":["math"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsNestedArrayItemsWithKeyCleanupScalarAndNesting(t *testing.T) {
	const zwsp = "\u200b"

	input := `{people:[
	  {"na` + zwsp + `me":'Ada',event:{date:'28 May 2026',city:'Paris',country:'France'},score:'1.5',status:'na'},
	  {name:'Grace',event:{date:'2026/05/29',city:'London',country:'UK'},score:'2',status:'ready',},
	]}`
	want := `{"people":[{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":1.5,"status":null},{"event":{"date":"2026-05-29","location":{"city":"London","country":"UK"}},"name":"Grace","score":2,"status":"ready"}]}`

	assertEnsurePipeline(t, input, pipelinePeopleSchema, want)
}

func TestEnsureFullPipelineRepairsOneOfBranchAfterWrapperUnwrapAndScalarOutput(t *testing.T) {
	input := `{"answer":{"count":"42"}}`
	want := `{"count":42}`

	assertEnsurePipeline(t, input, pipelineOneOfCountOrDateSchema, want)
}

func TestEnsureFullPipelineRepairsFencedRelaxedWrappedEnumOutput(t *testing.T) {
	input := "Here is the value:\n```json\n{payload:{status:'IN_PROGRESS'}}\n```"
	want := `{"status":"in-progress"}`

	assertEnsurePipeline(t, input, statusEnumSchema, want)
}

func TestEnsureFullPipelineRepeatsSchemaStageForNestingAndMultipleScalars(t *testing.T) {
	input := `{"name":"Ada","event":{"date":"28 May 2026","city":"Paris","country":"France"},"score":"42.5","status":"not available","tags":["research"]}`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsInverseNestingWithNestedKeyCleanupAndScalarOutput(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"location":{"ci` + zwsp + `ty":"Paris","country":"France"},"date":"28/05/2026","count":"7","status":"--"}`
	want := `{"city":"Paris","count":7,"country":"France","date":"2026-05-28","status":null}`

	assertEnsurePipeline(t, input, pipelineFlatEventSchema, want)
}

func TestEnsureFullPipelineRepairsInverseNestingAfterResponseEnvelopeUnwrap(t *testing.T) {
	input := `{"answer":{"location":{"city":"Paris","country":"France"},"date":"28 May 2026","count":"7","status":"N/A"}}`
	want := `{"city":"Paris","count":7,"country":"France","date":"2026-05-28","status":null}`

	assertEnsurePipeline(t, input, pipelineFlatEventSchema, want)
}

func TestEnsureFullPipelinePreservesZeroWidthValuesWhileOtherRepairsRun(t *testing.T) {
	const zwsp = "\u200b"

	input := `{
	  response: {
	    "na` + zwsp + `me": "A` + zwsp + `da",
	    event: {
	      date: '28 May 2026',
	      city: 'Paris',
	      country: 'France'
	    },
	    score: "1",
	    status: "ready",
	    tags: ["ke` + zwsp + `ep"]
	  }
	}`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"A` + zwsp + `da","score":1,"status":"ready","tags":["ke` + zwsp + `ep"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRejectsUnrecoverableInvalidJSON(t *testing.T) {
	_, err := schema_compliance.Ensure("```json\n{name: Ada Lovelace}\n```", basicObjectSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindInvalidJSON)
}

func TestEnsureFullPipelineRejectsRemainingSchemaViolation(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"data":{"name":42}}`, basicObjectSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRejectsAmbiguousDateAfterOtherRepairs(t *testing.T) {
	input := `{"data":{"name":"Ada","event":{"date":"05/06/2026","city":"Paris","country":"France"},"score":"1","status":"none","tags":["research"]}}`
	_, err := schema_compliance.Ensure(input, pipelineProfileSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRejectsUnsafeMultiKeyResponseEnvelope(t *testing.T) {
	input := `{"data":{"name":"Ada"},"meta":{"request_id":"1"}}`
	_, err := schema_compliance.Ensure(input, basicObjectSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRejectsZeroWidthKeyCollision(t *testing.T) {
	_, err := schema_compliance.Ensure("```json\n{data:{name:'Ada',\"na\u200bme\":'Grace'}}\n```", basicObjectSchema)
	assertSchemaViolationError(t, err)
}

func TestEnsureFullPipelineRejectsNestedZeroWidthKeyCollision(t *testing.T) {
	_, err := schema_compliance.Ensure(`{"people":[{"name":"Ada","na\u200bme":"Grace"}]}`, pipelinePeopleSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func assertEnsurePipeline(t *testing.T, input string, schema string, want string) {
	t.Helper()

	got, err := schema_compliance.Ensure(input, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Ensure() = %q, want %q", got, want)
	}
}

func assertEnsurePipelineErrorKind(t *testing.T, err error, want schema_compliance.ErrorKind) {
	t.Helper()
	if err == nil {
		t.Fatal("Ensure returned nil error")
	}

	var complianceErr *schema_compliance.Error
	if !errors.As(err, &complianceErr) {
		t.Fatalf("error type = %T, want *schema_compliance.Error", err)
	}
	if complianceErr.Kind != want {
		t.Fatalf("error kind = %v, want %v", complianceErr.Kind, want)
	}
}
