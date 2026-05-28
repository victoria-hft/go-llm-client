// Package schema_compliance converts LLM JSON-like output into schema-compliant JSON.
package schema_compliance

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaResource = "schema.json"

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
		extractSurroundedFencedJSON,
	}
}

func iterativeFixes() []fixFunc {
	return nil
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
