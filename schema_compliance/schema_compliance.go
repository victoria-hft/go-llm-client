// Package schema_compliance converts LLM JSON-like output into schema-compliant JSON.
package schema_compliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaResource = "schema.json"

// ErrorKind identifies why Ensure could not return schema-compliant JSON.
type ErrorKind string

const (
	// ErrorKindInvalidJSON indicates the final output could not be parsed as JSON.
	ErrorKindInvalidJSON ErrorKind = "invalid_json"

	// ErrorKindInvalidSchema indicates the supplied schema could not be parsed or compiled.
	ErrorKindInvalidSchema ErrorKind = "invalid_schema"

	// ErrorKindSchemaViolation indicates the JSON did not match the supplied schema.
	ErrorKindSchemaViolation ErrorKind = "schema_violation"
)

// Error is returned for all schema compliance failures.
type Error struct {
	Kind ErrorKind
	Err  error
}

// Error returns a human-readable schema compliance error.
func (e *Error) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}
	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

// Unwrap returns the wrapped lower-level error, if any.
func (e *Error) Unwrap() error {
	return e.Err
}

// Ensure returns valid JSON that conforms to schemaJSON, or a typed error.
func Ensure(output string, schemaJSON string) (string, error) {
	schema, err := compileSchema(schemaJSON)
	if err != nil {
		return "", &Error{Kind: ErrorKindInvalidSchema, Err: err}
	}

	current := output
	for _, fix := range oneTimeFixes() {
		next, changed := fix(current)
		if changed {
			current = next
		}
	}

	for {
		changed := false
		for _, fix := range iterativeFixes() {
			next, didChange := fix(current)
			if didChange {
				current = next
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}

	value, normalized, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", &Error{Kind: ErrorKindInvalidJSON, Err: err}
	}
	if err := schema.Validate(value); err != nil {
		return "", &Error{Kind: ErrorKindSchemaViolation, Err: err}
	}
	return normalized, nil
}

type fixFunc func(string) (string, bool)

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		extractFencedJSONBlock,
	}
}

func iterativeFixes() []fixFunc {
	return nil
}

func extractFencedJSONBlock(output string) (string, bool) {
	start := strings.Index(output, "```")
	if start == -1 {
		return output, false
	}

	bodyStart := start + len("```")
	end := strings.Index(output[bodyStart:], "```")
	if end == -1 {
		return output, false
	}
	end += bodyStart

	body := strings.TrimSpace(output[bodyStart:end])
	if body == "" {
		return output, false
	}

	if startsWithJSON(body) {
		withoutLanguage := strings.TrimSpace(body[len("json"):])
		if withoutLanguage != "" {
			body = withoutLanguage
		}
	}

	if body == "" {
		return output, false
	}
	return body, true
}

func startsWithJSON(body string) bool {
	if len(body) < len("json") || !strings.EqualFold(body[:len("json")], "json") {
		return false
	}
	return len(body) == len("json") || isWhitespace(body[len("json")])
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}

func compileSchema(schemaJSON string) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaJSON))
	if err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaResource, doc); err != nil {
		return nil, err
	}
	return compiler.Compile(schemaResource)
}

func parseAndNormalizeJSON(output string) (any, string, error) {
	trimmed := strings.TrimSpace(output)
	value, err := jsonschema.UnmarshalJSON(strings.NewReader(trimmed))
	if err != nil {
		return nil, "", err
	}

	var normalized bytes.Buffer
	if err := compactJSON(&normalized, trimmed); err != nil {
		return nil, "", err
	}
	return value, normalized.String(), nil
}

func compactJSON(dst *bytes.Buffer, src string) error {
	return json.Compact(dst, []byte(src))
}
