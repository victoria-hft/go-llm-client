package schema_compliance

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repairNDJSONArrayOutput(current string, schema *jsonschema.Schema) (string, bool) {
	if !schemaHasArrayBranch(schema) {
		return current, false
	}

	lines := nonEmptyNDJSONLines(current)
	if len(lines) < 2 {
		return current, false
	}

	values := make([]any, 0, len(lines))
	for _, line := range lines {
		value, _, err := parseAndNormalizeJSON(line)
		if err != nil {
			return current, false
		}
		values = append(values, value)
	}

	repaired, err := marshalCanonicalJSON(values)
	if err != nil {
		return current, false
	}
	if repaired == strings.TrimSpace(current) {
		return current, false
	}
	return repaired, true
}

func schemaHasArrayBranch(schema *jsonschema.Schema) bool {
	for _, branch := range candidateSchemaBranches(schema) {
		if schemaExpectsArray(branch) {
			return true
		}
	}
	return false
}

func nonEmptyNDJSONLines(input string) []string {
	rawLines := strings.Split(strings.TrimSpace(input), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, rawLine := range rawLines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
