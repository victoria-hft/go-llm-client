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

const pipelineTextTagsSchema = `{
  "type": "object",
  "required": ["text", "tags"],
  "properties": {
    "text": {"type": "string"},
    "tags": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "additionalProperties": false
}`

const pipelineStringMapSchema = `{
  "type": "object",
  "additionalProperties": {"type": "string"}
}`

const pipelineBuildResultSchema = `{
  "type": "object",
  "required": ["title", "status", "commands", "metadata", "note"],
  "properties": {
    "title": {"type": "string"},
    "status": {
      "type": "string",
      "enum": ["in-progress", "done"]
    },
    "commands": {
      "type": "array",
      "items": {"type": "string"}
    },
    "metadata": {
      "type": "object",
      "additionalProperties": {"type": "string"}
    },
    "note": {"type": "string"}
  },
  "additionalProperties": false
}`

const pipelineNestedBuildsSchema = `{
  "type": "object",
  "required": ["builds"],
  "properties": {
    "builds": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["title", "status", "commands", "metadata", "note"],
        "properties": {
          "title": {"type": "string"},
          "status": {
            "type": "string",
            "enum": ["in-progress", "done"]
          },
          "commands": {
            "type": "array",
            "items": {"type": "string"}
          },
          "metadata": {
            "type": "object",
            "additionalProperties": {"type": "string"}
          },
          "note": {"type": "string"}
        },
        "additionalProperties": false
      }
    }
  },
  "additionalProperties": false
}`

const pipelineTimedJobSchema = `{
  "type": "object",
  "required": ["name", "run", "status", "steps"],
  "properties": {
    "name": {"type": "string"},
    "run": {
      "type": "object",
      "required": ["date", "timestamp", "duration"],
      "properties": {
        "date": {"type": "string", "format": "date"},
        "timestamp": {"type": "string", "format": "date-time"},
        "duration": {"type": "string", "format": "duration"}
      },
      "additionalProperties": false
    },
    "status": {
      "type": "string",
      "enum": ["in-progress", "done"]
    },
    "steps": {
      "type": "array",
      "items": {"type": "string"}
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

func TestEnsureFullPipelineRepairsSmartQuotesAndNumericKeyArrayOutput(t *testing.T) {
	input := "Here is the value:\n```json\n{payload:{“text”: “plain value”, tags: {\"0\": 'a', \"1\": 'b'}}}\n```"
	want := `{"tags":["a","b"],"text":"plain value"}`

	assertEnsurePipeline(t, input, pipelineTextTagsSchema, want)
}

func TestEnsureFullPipelineRepairsRelaxedWrappedKeyValueArrayObjectOutput(t *testing.T) {
	input := "Here is the value:\n```\n{answer:[{name:'a', value:'one'}, {name:'b', value:'two'}]}\n```"
	want := `{"a":"one","b":"two"}`

	assertEnsurePipeline(t, input, pipelineStringMapSchema, want)
}

func TestEnsureFullPipelineRepairsBuildResultWithAllNewFixersAndExistingStages(t *testing.T) {
	input := "Here is the value:\n```json\n" +
		`{
		  payload: {
		    “title”: “Build”,
		    status: 'IN_PROGRESS',
		    commands: {“1”: “go test ./...”, “2”: “make”},
		    metadata: [{name:'os', value:'linux'}, {name:'arch', value:'arm64'}],
		    note: “escaped text”
		  }
		}` +
		"\n```"
	want := `{"commands":["go test ./...","make"],"metadata":{"arch":"arm64","os":"linux"},"note":"escaped text","status":"in-progress","title":"Build"}`

	assertEnsurePipeline(t, input, pipelineBuildResultSchema, want)
}

func TestEnsureFullPipelineRepairsNestedBuildsWithNumericArraysAndKeyValueObjects(t *testing.T) {
	input := `{
	  builds: {
	    "0": {
	      title: 'Unit',
	      status: 'DONE',
	      commands: {"0": "go test ./schema_compliance", "1": "make"},
	      metadata: [{key: 'suite', value: 'schema'}],
	      note: "unit checks"
	    },
	    "1": {
	      title: 'Lint',
	      status: 'in_progress',
	      commands: {"1": "go fmt ./...", "2": "go vet ./..."},
	      metadata: [{field: 'suite', value: 'lint'}],
	      note: "lint checks"
	    }
	  }
	}`
	want := `{"builds":[{"commands":["go test ./schema_compliance","make"],"metadata":{"suite":"schema"},"note":"unit checks","status":"done","title":"Unit"},{"commands":["go fmt ./...","go vet ./..."],"metadata":{"suite":"lint"},"note":"lint checks","status":"in-progress","title":"Lint"}]}`

	assertEnsurePipeline(t, input, pipelineNestedBuildsSchema, want)
}

func TestEnsureFullPipelineRepairsSmartQuotesBeforeRelaxedJSON(t *testing.T) {
	input := "```\n{“text”: “plain value”, tags: {“0”: 'math', “1”: 'proof'}}\n```"
	want := `{"tags":["math","proof"],"text":"plain value"}`

	assertEnsurePipeline(t, input, pipelineTextTagsSchema, want)
}

func TestEnsureFullPipelineRepairsTimeScalarsWithExistingStages(t *testing.T) {
	input := "Here is the value:\n```json\n" +
		`{
		  payload: {
		    name: 'Nightly',
		    run: {
		      date: '2026/05/28',
		      timestamp: 1779975900000,
		      duration: '5 minutes'
		    },
		    status: 'IN_PROGRESS',
		    steps: {“1”: 'go test ./...', “2”: 'make'}
		  }
		}` +
		"\n```"
	want := `{"name":"Nightly","run":{"date":"2026-05-28","duration":"PT5M","timestamp":"2026-05-28T13:45:00Z"},"status":"in-progress","steps":["go test ./...","make"]}`

	assertEnsurePipeline(t, input, pipelineTimedJobSchema, want)
}

func TestEnsureFullPipelineRepairsNestedTimeScalarsAfterKeyValueObjectConversion(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["jobs"],
	  "properties": {
	    "jobs": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["metadata", "timestamp", "duration"],
	        "properties": {
	          "metadata": {
	            "type": "object",
	            "additionalProperties": {"type": "string"}
	          },
	          "timestamp": {"type": "string", "format": "date-time"},
	          "duration": {"type": "string", "format": "duration"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{result:{jobs:{"0":{metadata:[{name:'owner', value:'Ada'}], timestamp:'2026-05-28 13:45:00z', duration:'2 hours'}}}}`
	want := `{"jobs":[{"duration":"PT2H","metadata":{"owner":"Ada"},"timestamp":"2026-05-28T13:45:00Z"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsKeyValueMapThenEnumAndDateScalars(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["properties", "status", "date"],
	  "properties": {
	    "properties": {
	      "type": "object",
	      "additionalProperties": {"type": "string"}
	    },
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "date": {"type": "string", "format": "date"}
	  },
	  "additionalProperties": false
	}`

	input := `{data:{properties:[{property:'owner', value:'Ada'}, {property:'team', value:'research'}], status:'IN_PROGRESS', date:'28 May 2026'}}`
	want := `{"date":"2026-05-28","properties":{"owner":"Ada","team":"research"},"status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNumericArrayThenNestedFieldAndScalarValues(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["events"],
	  "properties": {
	    "events": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["date", "location", "score"],
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
	          },
	          "score": {"type": "number"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{result:{events:{"1":{date:'28 May 2026', city:'Paris', country:'France', score:'9.5'}, "2":{date:'29 May 2026', city:'Lyon', country:'France', score:'8'}}}}`
	want := `{"events":[{"date":"2026-05-28","location":{"city":"Paris","country":"France"},"score":9.5},{"date":"2026-05-29","location":{"city":"Lyon","country":"France"},"score":8}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsKeyValueArrayInsideNumericKeyArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["rows"],
	  "properties": {
	    "rows": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["attributes"],
	        "properties": {
	          "attributes": {
	            "type": "object",
	            "additionalProperties": {"type": "string"}
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{output:{rows:{"0":{attributes:[{id:'a', value:'one'}]}, "1":{attributes:[{key:'b', val:'two'}]}}}}`
	want := `{"rows":[{"attributes":{"a":"one"}},{"attributes":{"b":"two"}}]}`

	assertEnsurePipeline(t, input, schema, want)
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
