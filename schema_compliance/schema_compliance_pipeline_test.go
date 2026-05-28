package schema_compliance_test

import (
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

func TestEnsureFullPipelineRepairsTruncatedRelaxedZeroWidthWrappedSchemaLoop(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"data":{"na` + zwsp + `me":'Ada',event:{date:'May 28, 2026',city:'Paris',country:'France'},score:"42.5",status:"unknown",tags:['research']`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsSingleItemArrayWrapperWithRelaxedNestedScalarOutput(t *testing.T) {
	input := `[{name:'Ada',event:{date:'2026/05/28',city:'Paris',country:'France'},score:'42',status:'null',tags:['math',]}]`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42,"status":null,"tags":["math"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsInverseNestingWithNestedKeyCleanupAndScalarOutput(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"location":{"ci` + zwsp + `ty":"Paris","country":"France"},"date":"28/05/2026","count":"7","status":"--"}`
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

func TestEnsureFullPipelineRejectsZeroWidthKeyCollision(t *testing.T) {
	_, err := schema_compliance.Ensure("```json\n{data:{name:'Ada',\"na\u200bme\":'Grace'}}\n```", basicObjectSchema)
	assertSchemaViolationError(t, err)
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
