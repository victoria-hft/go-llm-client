package schema_compliance_test

import (
	"testing"

	"github.com/victoria-hft/go-llm-client/schema_compliance"
)

const isoDateSchema = `{
  "type": "object",
  "required": ["date"],
  "properties": {
    "date": {
      "type": "string",
      "format": "date"
    }
  },
  "additionalProperties": false
}`

func TestEnsureRepairsHumanDateToISODate(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"date":"28 May 2026"}`, isoDateSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"date":"2026-05-28"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"date":"2026-05-28"}`)
	}
}

func TestEnsureRepairsConservativeDateFormats(t *testing.T) {
	tests := map[string]string{
		"28 May, 2026": "2026-05-28",
		"May 28 2026":  "2026-05-28",
		"May 28, 2026": "2026-05-28",
		"2026/05/28":   "2026-05-28",
		"28/05/2026":   "2026-05-28",
	}

	for input, wantDate := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"date":"`+input+`"}`, isoDateSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"date":"` + wantDate + `"}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsDateRecursivelyInObjectArray(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["events"],
	  "properties": {
	    "events": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "required": ["date"],
	        "properties": {
	          "date": {"type": "string", "format": "date"}
	        },
	        "additionalProperties": false
	      }
	    }
	  },
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"events":[{"date":"28 May 2026"}]}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"events":[{"date":"2026-05-28"}]}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"events":[{"date":"2026-05-28"}]}`)
	}
}

func TestEnsureRepairsScalarUsingOneOfBranch(t *testing.T) {
	const schema = `{
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

	got, err := schema_compliance.Ensure(`{"count":"42"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"count":42}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"count":42}`)
	}
}

func TestEnsureRepairsIntegerString(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["count"],
	  "properties": {"count": {"type": "integer"}},
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"count":"42"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"count":42}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"count":42}`)
	}
}

func TestEnsureRepairsNumberString(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["score"],
	  "properties": {"score": {"type": "number"}},
	  "additionalProperties": false
	}`

	got, err := schema_compliance.Ensure(`{"score":"42.5"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"score":42.5}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"score":42.5}`)
	}
}

func TestEnsureDoesNotRepairFloatStringForIntegerSchema(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["count"],
	  "properties": {"count": {"type": "integer"}},
	  "additionalProperties": false
	}`

	_, err := schema_compliance.Ensure(`{"count":"42.5"}`, schema)
	assertSchemaViolationError(t, err)
}

func TestEnsureRepairsPlaceholderStringsToNull(t *testing.T) {
	const schema = `{
	  "type": "object",
	  "required": ["status"],
	  "properties": {
	    "status": {
	      "type": ["string", "null"],
	      "enum": ["ready", "done", null]
	    }
	  },
	  "additionalProperties": false
	}`

	tests := []string{"N/A", "", "unknown", "null"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"status":"`+input+`"}`, schema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != `{"status":null}` {
				t.Fatalf("Ensure() = %q, want %q", got, `{"status":null}`)
			}
		})
	}
}

func TestEnsureDoesNotRepairPlaceholderWhenArbitraryStringIsAllowed(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"unknown"}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"unknown"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"unknown"}`)
	}
}
