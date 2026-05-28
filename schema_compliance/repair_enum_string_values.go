package schema_compliance

import (
	"strings"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repairEnumStringValues(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateEnumStringValueCandidates(value, schema, func(candidate any) bool {
		candidateJSON, err := marshalCanonicalJSON(candidate)
		if err != nil {
			return false
		}
		if schemaLoss(candidateJSON, schema) >= currentLoss {
			return false
		}
		repaired = candidateJSON
		return true
	})
	return repaired, found
}

func enumerateEnumStringValueCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if text, ok := value.(string); ok {
		for _, branch := range branches {
			if candidate, ok := enumStringValueCandidate(text, branch); ok {
				if yield(candidate) {
					return true
				}
			}
		}
		return false
	}

	if object, ok := value.(map[string]any); ok {
		for _, key := range sortedObjectKeys(object) {
			child := object[key]
			for _, branch := range branches {
				propertySchema, ok := branch.Properties[key]
				if !ok {
					continue
				}
				if enumerateEnumStringValueCandidates(child, propertySchema, func(candidateChild any) bool {
					candidate := cloneJSONObject(object)
					candidate[key] = candidateChild
					return yield(candidate)
				}) {
					return true
				}
			}
		}
		return false
	}

	if array, ok := value.([]any); ok {
		for _, branch := range branches {
			for index, item := range array {
				for _, itemSchema := range itemSchemasForIndex(arrayItemSchemas(branch), branch, index) {
					if enumerateEnumStringValueCandidates(item, itemSchema, func(candidateItem any) bool {
						candidate := cloneJSONArray(array)
						candidate[index] = candidateItem
						return yield(candidate)
					}) {
						return true
					}
				}
			}
		}
	}

	return false
}

func enumStringValueCandidate(value string, schema *jsonschema.Schema) (string, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil || schema.Enum == nil {
		return "", false
	}

	normalizedValue := normalizeEnumString(value)
	if normalizedValue == "" {
		return "", false
	}

	matches := enumStringMatchesByNormalizedValue(schema.Enum.Values, normalizedValue)
	if len(matches) != 1 || matches[0] == value {
		return "", false
	}
	return matches[0], true
}

func enumStringMatchesByNormalizedValue(values []any, normalizedValue string) []string {
	var matches []string
	seen := map[string]bool{}
	for _, enumValue := range values {
		text, ok := enumValue.(string)
		if !ok {
			continue
		}
		if normalizeEnumString(text) != normalizedValue {
			continue
		}
		if seen[text] {
			continue
		}
		seen[text] = true
		matches = append(matches, text)
	}
	return matches
}

func normalizeEnumString(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r == '_' || r == '-' || unicode.IsSpace(r) {
			continue
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}
