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

const pipelineNumericMetricsSchema = `{
  "type": "object",
  "required": ["name", "metrics", "status"],
  "properties": {
    "name": {"type": "string"},
    "metrics": {
      "type": "object",
      "required": ["count", "ratio", "rate", "spread", "limit"],
      "properties": {
        "count": {"type": "integer"},
        "ratio": {"type": "number", "minimum": 0, "maximum": 1},
        "rate": {"type": "number"},
        "spread": {"type": "number", "minimum": 0, "maximum": 1},
        "limit": {"type": "integer"}
      },
      "additionalProperties": false
    },
    "status": {
      "type": "string",
      "enum": ["in-progress", "done"]
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

func TestEnsureFullPipelineRepairsFencedRelaxedWrappedEnumExplanationOutput(t *testing.T) {
	input := "Here is the value:\n```json\n{payload:{status:'In Progress: currently active'}}\n```"
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

func TestEnsureFullPipelineRepairsNumericScalarsWithExistingStages(t *testing.T) {
	input := "Here is the value:\n```json\n" +
		`{
		  payload: {
		    name: 'Metrics',
		    status: 'IN_PROGRESS',
		    metrics: {
		      count: '1_000',
		      ratio: '92%',
		      rate: '1e-6',
		      spread: '25 bps',
		      limit: 0xFF
		    }
		  }
		}` +
		"\n```"
	want := `{"metrics":{"count":1000,"limit":255,"rate":1e-6,"ratio":0.92,"spread":0.0025},"name":"Metrics","status":"in-progress"}`

	assertEnsurePipeline(t, input, pipelineNumericMetricsSchema, want)
}

func TestEnsureFullPipelineRepairsFencedRelaxedMissingHTTPSURL(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["source_url", "status"],
	  "properties": {
	    "source_url": {"type": "string"},
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    }
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```\n{source_url:'x.com/a b', status:'IN_PROGRESS'}\n```"
	want := `{"source_url":"https://x.com/a%20b","status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsWrappedURIWhitespaceWithDateAndNumber(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["url", "date", "score"],
	  "properties": {
	    "url": {"type": "string", "format": "uri"},
	    "date": {"type": "string", "format": "date"},
	    "score": {"type": "number"}
	  },
	  "additionalProperties": false
	}`

	input := `{payload:{url:'https://x.com/report path?q=a b', date:'28 May 2026', score:'3/4'}}`
	want := `{"date":"2026-05-28","score":0.75,"url":"https://x.com/report%20path?q=a%20b"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNumericKeyArrayItemsWithURLs(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["links"],
	  "properties": {
	    "links": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["url", "active"],
	        "properties": {
	          "url": {"type": "string", "format": "uri"},
	          "active": {"type": "boolean"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{result:{links:{"1":{url:'x.com/a b', active:"\u2705"}, "2":{url:'example.com/c d', active:"\u274c"}}}}`
	want := `{"links":[{"active":true,"url":"https://x.com/a%20b"},{"active":false,"url":"https://example.com/c%20d"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsSmartQuotesZeroWidthURLKeyAndEnumExplanation(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["source_url", "sentiment"],
	  "properties": {
	    "source_url": {"type": "string"},
	    "sentiment": {
	      "type": "string",
	      "enum": ["positive", "neutral", "negative"]
	    }
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := "```json\n{“source_" + zwsp + "url”: “x.com/café path”, sentiment:'Positive: customer is satisfied'}\n```"
	want := `{"sentiment":"positive","source_url":"https://x.com/caf%C3%A9%20path"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsFencedWrappedUUIDWithEnumAndScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["id", "status", "date"],
	  "properties": {
	    "id": {"type": "string"},
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "date": {"type": "string", "format": "date"}
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```json\n{payload:{id:'F0307D30EAD2417392873DAB7FFA0FA4', status:'IN_PROGRESS', date:'28 May 2026'}}\n```"
	want := `{"date":"2026-05-28","id":"f0307d30-ead2-4173-9287-3dab7ffa0fa4","status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsUUIDsInsideNumericKeyArrayItems(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["records"],
	  "properties": {
	    "records": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["id", "source_url"],
	        "properties": {
	          "id": {"type": "string"},
	          "source_url": {"type": "string"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := `{result:{records:{"1":{"id":"urn:uuid:F0307D30-EAD2-4173-9287-3DAB7FFA0FA4","source_` + zwsp + `url":"x.com/a b"},"2":{"id":"f0307d30ead2-41739287-3dab7ffa0fa4","source_url":"example.com/c d"}}}}`
	want := `{"records":[{"id":"f0307d30-ead2-4173-9287-3dab7ffa0fa4","source_url":"https://x.com/a%20b"},{"id":"f0307d30-ead2-4173-9287-3dab7ffa0fa4","source_url":"https://example.com/c%20d"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsInvalidUUIDWhenFormatRequiresUUID(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["id"],
	  "properties": {
	    "id": {"type": "string", "format": "uuid"}
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{payload:{id:'f0307d30--ead2-4173-9287-3dab7ffa0fa4'}}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsItemItemsShapeWithURLAndScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["url", "score"],
	        "properties": {
	          "url": {"type": "string", "format": "uri"},
	          "score": {"type": "number"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := "```json\n{item:{url:'x.com/a b', score:'1_000'}}\n```"
	want := `{"items":[{"score":1000,"url":"https://x.com/a%20b"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsURLRepairWhenSchemaDoesNotAllowIt(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["target"],
	  "properties": {
	    "target": {"type": "string", "pattern": "^https://"}
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{payload:{target:'x.com/a b'}}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsFencedNDJSONWithScalarAndEnumFixes(t *testing.T) {
	const schema = `{
	  "type": "array",
	  "items": {
	    "type": "object",
	    "required": ["date", "score", "status"],
	    "properties": {
	      "date": {"type": "string", "format": "date"},
	      "score": {"type": "number"},
	      "status": {"type": "string", "enum": ["ready", "done"]}
	    },
	    "additionalProperties": false
	  }
	}`
	input := "Rows:\n```json\n" +
		"{\"date\":\"28 May 2026\",\"score\":\"42.5\",\"status\":\"READY\"}\n" +
		"{\"date\":\"2026/05/29\",\"score\":\"3/4\",\"status\":\"Done: complete\"}\n" +
		"```\n"
	want := `[{"date":"2026-05-28","score":42.5,"status":"ready"},{"date":"2026-05-29","score":0.75,"status":"done"}]`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNDJSONItemsWithNumericKeyArrays(t *testing.T) {
	const schema = `{
	  "type": "array",
	  "items": {
	    "type": "object",
	    "required": ["name", "steps"],
	    "properties": {
	      "name": {"type": "string"},
	      "steps": {
	        "type": "array",
	        "items": {"type": "string"}
	      }
	    },
	    "additionalProperties": false
	  }
	}`
	input := "{\"name\":\"build\",\"steps\":{\"0\":\"go test\",\"1\":\"make\"}}\n" +
		"{\"name\":\"deploy\",\"steps\":{\"1\":\"plan\",\"2\":\"apply\"}}"
	want := `[{"name":"build","steps":["go test","make"]},{"name":"deploy","steps":["plan","apply"]}]`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNDJSONItemsWithKeyCleanupAndURLRepair(t *testing.T) {
	const zwsp = "\u200b"
	const schema = `{
	  "type": "array",
	  "items": {
	    "type": "object",
	    "required": ["source_url", "title"],
	    "properties": {
	      "source_url": {"type": "string"},
	      "title": {"type": "string"}
	    },
	    "additionalProperties": false
	  }
	}`
	input := "{\"source_" + zwsp + "url\":\"x.com/a b\",\"title\":\"A\"}\n" +
		"{\"source_url\":\"example.com/c d\",\"title\":\"B\"}"
	want := `[{"source_url":"https://x.com/a%20b","title":"A"},{"source_url":"https://example.com/c%20d","title":"B"}]`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNDJSONItemsWithItemItemsAndNesting(t *testing.T) {
	const schema = `{
	  "type": "array",
	  "items": {
	    "type": "object",
	    "required": ["items", "event"],
	    "properties": {
	      "items": {
	        "type": "array",
	        "items": {"type": "string"}
	      },
	      "event": {
	        "type": "object",
	        "required": ["location"],
	        "properties": {
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
	      }
	    },
	    "additionalProperties": false
	  }
	}`
	input := "{\"item\":\"alpha\",\"event\":{\"city\":\"Paris\",\"country\":\"France\"}}\n" +
		"{\"item\":\"beta\",\"event\":{\"city\":\"London\",\"country\":\"UK\"}}"
	want := `[{"event":{"location":{"city":"Paris","country":"France"}},"items":["alpha"]},{"event":{"location":{"city":"London","country":"UK"}},"items":["beta"]}]`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsNDJSONWhenOneItemCannotBecomeCompliant(t *testing.T) {
	input := "{\"date\":\"28 May 2026\",\"score\":\"42.5\",\"status\":\"READY\"}\n" +
		"{\"date\":\"05/06/2026\",\"score\":\"not numeric\",\"status\":\"ready\"}"

	_, err := schema_compliance.Ensure(input, `{
	  "type": "array",
	  "items": {
	    "type": "object",
	    "required": ["date", "score", "status"],
	    "properties": {
	      "date": {"type": "string", "format": "date"},
	      "score": {"type": "number"},
	      "status": {"type": "string", "enum": ["ready", "done"]}
	    },
	    "additionalProperties": false
	  }
	}`)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsFencedRelaxedUndefinedOutput(t *testing.T) {
	input := "Here is the value:\n```json\n{name:'Ada',event:{date:'28 May 2026',city:'Paris',country:'France'},score:'42.5',status:undefined,tags:['research']}\n```"
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsWrappedUndefinedWithScalarAndEnumOutput(t *testing.T) {
	input := `{answer:{city:'Paris',country:'France',date:'28 May 2026',count:'7',status:undefined}}`
	want := `{"city":"Paris","count":7,"country":"France","date":"2026-05-28","status":null}`

	assertEnsurePipeline(t, input, pipelineFlatEventSchema, want)
}

func TestEnsureFullPipelineRepairsUndefinedAfterZeroWidthKeyCleanup(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"na` + zwsp + `me":"Ada","event":{"date":"2026/05/28","city":"Paris","country":"France"},"score":"42","status":undefined,"tags":["research"]}`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsUndefinedWithItemItemsAndNesting(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items", "event", "status"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    },
	    "event": {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
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
	    "status": {"type": ["string", "null"], "enum": ["ready", null]}
	  },
	  "additionalProperties": false
	}`
	input := `{item:'alpha',event:{city:'Paris',country:'France'},status:undefined}`
	want := `{"event":{"location":{"city":"Paris","country":"France"}},"items":["alpha"],"status":null}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsFencedBareArrayWithEnumAndScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags", "status", "score"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string"}
	    },
	    "status": {"type": "string", "enum": ["in-progress", "done"]},
	    "score": {"type": "number"}
	  },
	  "additionalProperties": false
	}`
	input := "Here is the value:\n```json\n{tags:[risk, pricing], status:'IN_PROGRESS', score:'3/4'}\n```"
	want := `{"score":0.75,"status":"in-progress","tags":["risk","pricing"]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsFencedWrappedEnumArrayStringWithScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags", "status", "score"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["risk", "pricing", "legal"]}
	    },
	    "status": {"type": "string", "enum": ["in-progress", "done"]},
	    "score": {"type": "number"}
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```json\n{payload:{tags:'risk, pricing', status:'IN_PROGRESS', score:'3/4'}}\n```"
	want := `{"score":0.75,"status":"in-progress","tags":["risk","pricing"]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEnumArrayStringAfterZeroWidthKeyCleanup(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["risk", "pricing", "legal"]}
	    }
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := `{"ta` + zwsp + `gs":"risk,pricing"}`
	want := `{"tags":["risk","pricing"]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEnumArrayStringWithItemItemsAndNesting(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items", "event"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["tags"],
	        "properties": {
	          "tags": {
	            "type": "array",
	            "items": {"type": "string", "enum": ["risk", "pricing", "legal"]}
	          }
	        },
	        "additionalProperties": false
	      }
	    },
	    "event": {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
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
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{item:{tags:'risk,pricing'},event:{city:'Paris',country:'France'}}`
	want := `{"event":{"location":{"city":"Paris","country":"France"}},"items":[{"tags":["risk","pricing"]}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsEnumArrayStringWithUnknownToken(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["tags"],
	  "properties": {
	    "tags": {
	      "type": "array",
	      "items": {"type": "string", "enum": ["risk", "pricing", "legal"]}
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{payload:{tags:'risk,unknown'}}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsWrappedBareArrayOutput(t *testing.T) {
	input := `{answer:{text:'case',tags:[risk, pricing]}}`
	want := `{"tags":["risk","pricing"],"text":"case"}`

	assertEnsurePipeline(t, input, pipelineTextTagsSchema, want)
}

func TestEnsureFullPipelineRepairsBareArrayAfterZeroWidthKeyCleanup(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"te` + zwsp + `xt":"case",tags:[risk, pricing]}`
	want := `{"tags":["risk","pricing"],"text":"case"}`

	assertEnsurePipeline(t, input, pipelineTextTagsSchema, want)
}

func TestEnsureFullPipelineRejectsUnsafeBareArrayToken(t *testing.T) {
	_, err := schema_compliance.Ensure("```json\n{text:'case',tags:[high risk, pricing]}\n```", pipelineTextTagsSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindInvalidJSON)
}

func TestEnsureFullPipelineRepairsFencedWrappedNaNNullableNumber(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["score", "status"],
	  "properties": {
	    "score": {"type": ["number", "null"]},
	    "status": {"type": "string", "enum": ["in-progress", "done"]}
	  },
	  "additionalProperties": false
	}`
	input := "Here is the value:\n```json\n{payload:{score:NaN,status:'IN_PROGRESS'}}\n```"
	want := `{"score":null,"status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNaNAfterZeroWidthKeyCleanup(t *testing.T) {
	const zwsp = "\u200b"
	const schema = `{
	  "type": "object",
	  "required": ["score"],
	  "properties": {
	    "score": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`
	input := `{"sco` + zwsp + `re":NaN}`
	want := `{"score":null}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNaNWithItemItemsAndNesting(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items", "event", "score"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    },
	    "event": {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
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
	    "score": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`
	input := `{item:'alpha',event:{city:'Paris',country:'France'},score:NaN}`
	want := `{"event":{"location":{"city":"Paris","country":"France"}},"items":["alpha"],"score":null}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsNaNForRequiredNumber(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["score"],
	  "properties": {
	    "score": {"type": "number"}
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{score:NaN}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsFencedPythonLiteralsWithRelaxedJSON(t *testing.T) {
	input := "Here is the value:\n```json\n{name:'Ada',event:{date:'28 May 2026',city:'Paris',country:'France'},score:'42.5',status:None,tags:['research']}\n```"
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42.5,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsWrappedPythonLiteralsWithScalarAndEnumOutput(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["sentiment", "date", "score", "ok"],
	  "properties": {
	    "sentiment": {
	      "type": "string",
	      "enum": ["positive", "neutral", "negative"]
	    },
	    "date": {"type": "string", "format": "date"},
	    "score": {"type": "number"},
	    "ok": {"type": "boolean"}
	  },
	  "additionalProperties": false
	}`
	input := `{answer:{sentiment:'Positive: customer is satisfied', date:'28 May 2026', score:'3/4', ok:True}}`
	want := `{"date":"2026-05-28","ok":true,"score":0.75,"sentiment":"positive"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsPythonLiteralsAfterZeroWidthKeyCleanup(t *testing.T) {
	const zwsp = "\u200b"

	input := `{"na` + zwsp + `me":"Ada","event":{"date":"2026/05/28","city":"Paris","country":"France"},"score":"42","status":None,"tags":["research"]}`
	want := `{"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}},"name":"Ada","score":42,"status":null,"tags":["research"]}`

	assertEnsurePipeline(t, input, pipelineProfileSchema, want)
}

func TestEnsureFullPipelineRepairsPythonLiteralsWithItemItemsAndNesting(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items", "event", "active"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    },
	    "event": {
	      "type": "object",
	      "required": ["location"],
	      "properties": {
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
	    "active": {"type": "boolean"}
	  },
	  "additionalProperties": false
	}`
	input := `{item:'alpha',event:{city:'Paris',country:'France'},active:False}`
	want := `{"active":false,"event":{"location":{"city":"Paris","country":"France"}},"items":["alpha"]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEmptyObjectToNullWithWrapperEnumAndScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status", "score", "tags"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "score": {"type": ["number", "null"]},
	    "tags": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```json\n{payload:{status:'IN_PROGRESS', score:{}, tags:null}}\n```"
	want := `{"score":null,"status":"in-progress","tags":[]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEmptyArrayToNullAfterZeroWidthKeyCleanup(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["name", "note"],
	  "properties": {
	    "name": {"type": "string"},
	    "note": {"type": ["string", "null"]}
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := `{"na` + zwsp + `me":"Ada","note":[]}`
	want := `{"name":"Ada","note":null}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNullToEmptyArrayWithNestingAndScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["event", "aliases"],
	  "properties": {
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
	    "aliases": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{answer:{event:{date:'28 May 2026',city:'Paris',country:'France'},aliases:null}}`
	want := `{"aliases":[],"event":{"date":"2026-05-28","location":{"city":"Paris","country":"France"}}}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRejectsNonEmptyContainerNullabilityRepair(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["score"],
	  "properties": {
	    "score": {"type": ["number", "null"]}
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{payload:{score:{value:'1_000'}}}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRejectsPythonLiteralWhenConvertedValueViolatesSchema(t *testing.T) {
	_, err := schema_compliance.Ensure("```json\n{name:None}\n```", basicObjectSchema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
}

func TestEnsureFullPipelineRepairsEnumExplanationWithScalarAndWrapperFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["sentiment", "date", "score"],
	  "properties": {
	    "sentiment": {
	      "type": "string",
	      "enum": ["positive", "neutral", "negative"]
	    },
	    "date": {"type": "string", "format": "date"},
	    "score": {"type": "number"}
	  },
	  "additionalProperties": false
	}`

	input := `{answer:{sentiment:'Positive: customer is satisfied', date:'28 May 2026', score:'3/4'}}`
	want := `{"date":"2026-05-28","score":0.75,"sentiment":"positive"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEnumExplanationsInsideNumericKeyArrayItems(t *testing.T) {
	input := `{result:{reviews:{"1":{sentiment:'Positive: customer is satisfied', score:'92%'}, "2":{sentiment:'negative — customer is unhappy', score:'25 bps'}}}}`
	const boundedSchema = `{
	  "type": "object",
	  "required": ["reviews"],
	  "properties": {
	    "reviews": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["sentiment", "score"],
	        "properties": {
	          "sentiment": {
	            "type": "string",
	            "enum": ["positive", "neutral", "negative"]
	          },
	          "score": {"type": "number", "minimum": 0, "maximum": 1}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`
	want := `{"reviews":[{"score":0.92,"sentiment":"positive"},{"score":0.0025,"sentiment":"negative"}]}`

	assertEnsurePipeline(t, input, boundedSchema, want)
}

func TestEnsureFullPipelineRepairsEnumExplanationWithItemItemsAndBooleanMarker(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["items"],
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["sentiment", "active"],
	        "properties": {
	          "sentiment": {
	            "type": "string",
	            "enum": ["positive", "neutral", "negative"]
	          },
	          "active": {"type": "boolean"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := "```json\n{item:{sentiment:'customer is satisfied — Positive', active:\"\\u2705\"}}\n```"
	want := `{"items":[{"active":true,"sentiment":"positive"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEnumExplanationWithSmartQuotesAndKeyCleanup(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["sentiment", "note"],
	  "properties": {
	    "sentiment": {
	      "type": "string",
	      "enum": ["positive", "neutral", "negative"]
	    },
	    "note": {"type": "string"}
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := "```json\n{“senti" + zwsp + "ment”: “POSITIVE. customer is satisfied”, note:'kept'}\n```"
	want := `{"note":"kept","sentiment":"positive"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsItemItemsShapeWithExistingStages(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status", "collections"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "collections": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["items", "score"],
	        "properties": {
	          "items": {
	            "type": "array",
	            "items": {"type": "string"}
	          },
	          "score": {"type": "number"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```json\n{payload:{status:'IN_PROGRESS', collections:{\"0\":{item:'Ada', score:'1_000'}}}}\n```"
	want := `{"collections":[{"items":["Ada"],"score":1000}],"status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNullItemToItemsWithFencedRelaxedAndEnum(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status", "items"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "items": {
	      "type": "array",
	      "items": {"type": "string"}
	    }
	  },
	  "additionalProperties": false
	}`

	input := "\ufeffHere is the value:\n```json\n{answer:{status:'IN_PROGRESS', item:null,}}\n```"
	want := `{"items":[],"status":"in-progress"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsEmptyItemsToNullableItemWithSmartQuotesAndWrapper(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status", "item"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["ready", "done"]
	    },
	    "item": {"type": ["string", "null"]}
	  },
	  "additionalProperties": false
	}`

	input := "```json\n{payload:{“status”: “READY”, “items”: [],}}\n```"
	want := `{"item":null,"status":"ready"}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsObjectItemsToItemWithNestedScalarFixes(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["item"],
	  "properties": {
	    "item": {
	      "type": "object",
	      "required": ["date", "score", "status"],
	      "properties": {
	        "date": {"type": "string", "format": "date"},
	        "score": {"type": "number"},
	        "status": {
	          "type": "string",
	          "enum": ["in-progress", "done"]
	        }
	      },
	      "additionalProperties": false
	    }
	  },
	  "additionalProperties": false
	}`

	input := `{items:{date:'28 May 2026', score:'3/4', status:'IN_PROGRESS'}}`
	want := `{"item":{"date":"2026-05-28","score":0.75,"status":"in-progress"}}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsNestedItemItemsWithZeroWidthKeyAndNumericArray(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["groups"],
	  "properties": {
	    "groups": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["items", "status"],
	        "properties": {
	          "items": {
	            "type": "array",
	            "items": {"type": "string"}
	          },
	          "status": {
	            "type": "string",
	            "enum": ["ready", "done"]
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`
	const zwsp = "\u200b"

	input := `{
	  result: {
	    groups: {
	      "0": {"it` + zwsp + `em": 'Ada', status: 'READY'},
	      "1": {item: 'Grace', status: 'DONE'}
	    }
	  }
	}`
	want := `{"groups":[{"items":["Ada"],"status":"ready"},{"items":["Grace"],"status":"done"}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsItemsArrayToItemAfterNumericKeyArrayAndTimeScalar(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["records"],
	  "properties": {
	    "records": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["item"],
	        "properties": {
	          "item": {
	            "type": "object",
	            "required": ["timestamp"],
	            "properties": {
	              "timestamp": {"type": "string", "format": "date-time"}
	            },
	            "additionalProperties": false
	          }
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := "Here is the value:\n```\n{output:{records:{\"1\":{items:[{timestamp:'2026-05-28 13:45:00z'}]}}}}\n```"
	want := `{"records":[{"item":{"timestamp":"2026-05-28T13:45:00Z"}}]}`

	assertEnsurePipeline(t, input, schema, want)
}

func TestEnsureFullPipelineRepairsBooleanMarkersWithItemItemsAndScalars(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status", "items"],
	  "properties": {
	    "status": {
	      "type": "string",
	      "enum": ["in-progress", "done"]
	    },
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["active", "score", "date"],
	        "properties": {
	          "active": {"type": "boolean"},
	          "score": {"type": "number"},
	          "date": {"type": "string", "format": "date"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	input := "\ufeffHere is the value:\n```json\n{payload:{status:'IN_PROGRESS', item:{active:\"\\u2705\", score:'1_000', date:'28 May 2026'}}}\n```"
	want := `{"items":[{"active":true,"date":"2026-05-28","score":1000}],"status":"in-progress"}`

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

func TestEnsureFullPipelineRejectsAmbiguousEnumExplanation(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["sentiment"],
	  "properties": {
	    "sentiment": {
	      "type": "string",
	      "enum": ["positive", "neutral", "negative"]
	    }
	  },
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure("```json\n{payload:{sentiment:'positive — negative'}}\n```", schema)
	assertEnsurePipelineErrorKind(t, err, schema_compliance.ErrorKindSchemaViolation)
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
