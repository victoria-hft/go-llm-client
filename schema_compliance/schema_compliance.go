// Package schema_compliance converts LLM JSON-like output into schema-compliant JSON.
package schema_compliance

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Ensure returns valid JSON that conforms to schemaJSON, or a typed error.
func Ensure(output string, schemaJSON string) (string, error) {
	schema, err := compileSchema(schemaJSON)
	if err != nil {
		return "", &Error{Kind: ErrorKindInvalidSchema, Err: err}
	}

	if normalized, err := normalizeIfSchemaCompliant(output, schema); err == nil {
		return normalized, nil
	}

	current := output
	for _, fix := range oneTimeFixes() {
		next, changed := fix(current)
		if changed {
			current = next
		}
	}

	current = runFixesUntilStable(current, jsonSyntaxFixes())
	if err := ValidateJSON(current); err != nil {
		return "", err
	}

	if normalized, err := normalizeIfSchemaCompliant(current, schema); err == nil {
		return normalized, nil
	}

	current = runSchemaFixesUntilStable(current, schema, schemaComplianceFixes())
	normalized, err := normalizeIfSchemaCompliant(current, schema)
	if err == nil {
		return normalized, nil
	}
	return "", err
}

type fixFunc func(string) (string, bool)
type schemaFixFunc func(string, *jsonschema.Schema) (string, bool)

func runFixesUntilStable(current string, fixes []fixFunc) string {
	for {
		changed := false
		for _, fix := range fixes {
			next, didChange := fix(current)
			if didChange {
				current = next
				changed = true
				break
			}
		}
		if !changed {
			return current
		}
	}
}

func runSchemaFixesUntilStable(current string, schema *jsonschema.Schema, fixes []schemaFixFunc) string {
	for {
		changed := false
		for _, fix := range fixes {
			next, didChange := fix(current, schema)
			if didChange {
				current = next
				changed = true
				break
			}
		}
		if !changed {
			return current
		}
	}
}

// ValidateJSON returns nil when output is syntactically valid JSON.
func ValidateJSON(output string) error {
	_, _, err := parseAndNormalizeJSON(output)
	if err != nil {
		return &Error{Kind: ErrorKindInvalidJSON, Err: err}
	}
	return nil
}

// ValidateAgainstSchema returns nil when output is valid JSON that conforms to schemaJSON.
func ValidateAgainstSchema(output string, schemaJSON string) error {
	schema, err := compileSchema(schemaJSON)
	if err != nil {
		return &Error{Kind: ErrorKindInvalidSchema, Err: err}
	}
	_, err = normalizeIfSchemaCompliant(output, schema)
	return err
}

func normalizeIfSchemaCompliant(output string, schema *jsonschema.Schema) (string, error) {
	value, normalized, err := parseAndNormalizeJSON(output)
	if err != nil {
		return "", &Error{Kind: ErrorKindInvalidJSON, Err: err}
	}
	if err := schema.Validate(value); err != nil {
		return "", &Error{Kind: ErrorKindSchemaViolation, Err: err}
	}
	return normalized, nil
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

func marshalCanonicalJSON(value any) (string, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSpace(output.String()), nil
}
