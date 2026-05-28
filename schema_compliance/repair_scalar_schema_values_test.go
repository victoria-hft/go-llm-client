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

const isoDateTimeSchema = `{
  "type": "object",
  "required": ["timestamp"],
  "properties": {
    "timestamp": {
      "type": "string",
      "format": "date-time"
    }
  },
  "additionalProperties": false
}`

const isoDurationSchema = `{
  "type": "object",
  "required": ["duration"],
  "properties": {
    "duration": {
      "type": "string",
      "format": "duration"
    }
  },
  "additionalProperties": false
}`

const integerValueSchema = `{
  "type": "object",
  "required": ["value"],
  "properties": {
    "value": {"type": "integer"}
  },
  "additionalProperties": false
}`

const numberValueSchema = `{
  "type": "object",
  "required": ["value"],
  "properties": {
    "value": {"type": "number"}
  },
  "additionalProperties": false
}`

const probabilityValueSchema = `{
  "type": "object",
  "required": ["value"],
  "properties": {
    "value": {"type": "number", "minimum": 0, "maximum": 1}
  },
  "additionalProperties": false
}`

const percentValueSchema = `{
  "type": "object",
  "required": ["value"],
  "properties": {
    "value": {"type": "number", "minimum": 0, "maximum": 100}
  },
  "additionalProperties": false
}`

const booleanValueSchema = `{
  "type": "object",
  "required": ["value"],
  "properties": {
    "value": {"type": "boolean"}
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

func TestEnsureRepairsEpochTimestampsToISODate(t *testing.T) {
	tests := map[string]string{
		`1779975900`:       "2026-05-28",
		`"1779975900000"`:  "2026-05-28",
		`1779975900000000`: "2026-05-28",
	}

	for input, wantDate := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"date":`+input+`}`, isoDateSchema)
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

func TestEnsureRepairsEpochTimestampsToISODateTime(t *testing.T) {
	tests := map[string]string{
		`1779975900`:       "2026-05-28T13:45:00Z",
		`"1779975900000"`:  "2026-05-28T13:45:00Z",
		`1779975900000000`: "2026-05-28T13:45:00Z",
	}

	for input, wantTimestamp := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"timestamp":`+input+`}`, isoDateTimeSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"timestamp":"` + wantTimestamp + `"}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsDateTimeSeparatorAndUTCMarker(t *testing.T) {
	tests := map[string]string{
		"2026-05-28 13:45:00Z": "2026-05-28T13:45:00Z",
		"2026-05-28T13:45:00z": "2026-05-28T13:45:00Z",
		"2026-05-28 13:45:00z": "2026-05-28T13:45:00Z",
	}

	for input, wantTimestamp := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"timestamp":"`+input+`"}`, isoDateTimeSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"timestamp":"` + wantTimestamp + `"}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsSimpleDurationText(t *testing.T) {
	tests := map[string]string{
		"5 minutes": "PT5M",
		"2 hours":   "PT2H",
		"3 days":    "P3D",
		"1 week":    "P1W",
	}

	for input, wantDuration := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"duration":"`+input+`"}`, isoDurationSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"duration":"` + wantDuration + `"}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureDoesNotRepairUnsafeEpochTimestamps(t *testing.T) {
	tests := []string{
		`-1`,
		`1.7799759e9`,
		`1779975900.5`,
		`32535216001000000`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := schema_compliance.Ensure(`{"timestamp":`+input+`}`, isoDateTimeSchema)
			assertSchemaViolationError(t, err)
		})
	}
}

func TestEnsureDoesNotRepairUnsafeDurationText(t *testing.T) {
	tests := []string{
		"1 hour 30 minutes",
		"about 5 minutes",
		"5m",
		"1.5 minutes",
		"5",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := schema_compliance.Ensure(`{"duration":"`+input+`"}`, isoDurationSchema)
			assertSchemaViolationError(t, err)
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

func TestEnsureRepairsNumericSeparatorStrings(t *testing.T) {
	tests := map[string]string{
		"1_000":    "1000",
		"1,000":    "1000",
		"1 000":    "1000",
		"1,000.25": "1000.25",
		"1_000.25": "1000.25",
	}

	for input, wantValue := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"value":"`+input+`"}`, numberValueSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"value":` + wantValue + `}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsScientificNotationStringForNumber(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"value":"1e-6"}`, numberValueSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":1e-6}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":1e-6}`)
	}
}

func TestEnsureRepairsHexIntegerLiterals(t *testing.T) {
	tests := map[string]string{
		`"0xFF"`:  "255",
		`0xFF`:    "255",
		`"-0x10"`: "-16",
	}

	for input, wantValue := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"value":`+input+`}`, integerValueSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			want := `{"value":` + wantValue + `}`
			if got != want {
				t.Fatalf("Ensure() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureRepairsBigIntSuffixStringForInteger(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"value":"123n"}`, integerValueSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":123}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":123}`)
	}
}

func TestEnsureRepairsFractionStringForNumber(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"value":"3/4"}`, numberValueSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":0.75}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":0.75}`)
	}
}

func TestEnsureRepairsPercentAndBasisPointStringsUsingSchemaBounds(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		schema string
		want   string
	}{
		{name: "percent probability", input: "92%", schema: probabilityValueSchema, want: `{"value":0.92}`},
		{name: "percent 0 to 100", input: "92%", schema: percentValueSchema, want: `{"value":92}`},
		{name: "bps probability", input: "25 bps", schema: probabilityValueSchema, want: `{"value":0.0025}`},
		{name: "bps 0 to 100", input: "25 bps", schema: percentValueSchema, want: `{"value":0.25}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"value":"`+tt.input+`"}`, tt.schema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Ensure() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureRejectsUnsafeNumericRepairs(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		schema string
	}{
		{name: "comma decimal", input: `"1,25"`, schema: numberValueSchema},
		{name: "bad grouping", input: `"10,00"`, schema: numberValueSchema},
		{name: "mixed separators", input: `"1,000_000"`, schema: numberValueSchema},
		{name: "scientific integer", input: `"1e6"`, schema: integerValueSchema},
		{name: "hex number", input: `"0xFF"`, schema: numberValueSchema},
		{name: "hex overflow", input: `"0x8000000000000000"`, schema: integerValueSchema},
		{name: "bigint overflow", input: `"9223372036854775808n"`, schema: integerValueSchema},
		{name: "zero denominator", input: `"3/0"`, schema: numberValueSchema},
		{name: "percent without bounds", input: `"92%"`, schema: numberValueSchema},
		{name: "bps without bounds", input: `"25 bps"`, schema: numberValueSchema},
		{name: "out of range percent probability", input: `"250%"`, schema: probabilityValueSchema},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := schema_compliance.Ensure(`{"value":`+tt.input+`}`, tt.schema)
			assertSchemaViolationError(t, err)
		})
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

func TestEnsureRepairsBooleanMarkerStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "check mark", input: `"\u2713"`, want: `{"value":true}`},
		{name: "ballot x", input: `"\u2717"`, want: `{"value":false}`},
		{name: "white check mark", input: `"\u2705"`, want: `{"value":true}`},
		{name: "cross mark", input: `"\u274c"`, want: `{"value":false}`},
		{name: "heavy check mark", input: `"\u2714"`, want: `{"value":true}`},
		{name: "heavy ballot x", input: `"\u2718"`, want: `{"value":false}`},
		{name: "ballot box checked", input: `"\u2611"`, want: `{"value":true}`},
		{name: "ballot box x", input: `"\u2612"`, want: `{"value":false}`},
		{name: "yes", input: `" YES "`, want: `{"value":true}`},
		{name: "no", input: `" no "`, want: `{"value":false}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := schema_compliance.Ensure(`{"value":`+tt.input+`}`, booleanValueSchema)
			if err != nil {
				t.Fatalf("Ensure returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Ensure() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureRepairsBooleanMarkerUsingOneOfBranch(t *testing.T) {
	const schema = `{
	  "oneOf": [
	    {
	      "type": "object",
	      "required": ["value"],
	      "properties": {"value": {"type": "boolean"}},
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

	got, err := schema_compliance.Ensure(`{"value":"\u2705"}`, schema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"value":true}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"value":true}`)
	}
}

func TestEnsureDoesNotRepairBooleanMarkerForStringSchema(t *testing.T) {
	got, err := schema_compliance.Ensure(`{"name":"\u2705"}`, basicObjectSchema)
	if err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}
	if got != `{"name":"\u2705"}` {
		t.Fatalf("Ensure() = %q, want %q", got, `{"name":"\u2705"}`)
	}
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
