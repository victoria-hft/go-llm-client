package schema_compliance

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repairEnumStringArrayValues(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateEnumStringArrayValueCandidates(value, schema, func(candidate any) bool {
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

func enumerateEnumStringArrayValueCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if text, ok := value.(string); ok {
		for _, branch := range branches {
			if candidate, ok := enumStringArrayValueCandidate(text, branch); ok {
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
				if enumerateEnumStringArrayValueCandidates(child, propertySchema, func(candidateChild any) bool {
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
					if enumerateEnumStringArrayValueCandidates(item, itemSchema, func(candidateItem any) bool {
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

func enumStringArrayValueCandidate(value string, schema *jsonschema.Schema) ([]any, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil || !schemaExpectsArray(schema) {
		return nil, false
	}

	for _, itemSchema := range arrayItemSchemas(schema) {
		if candidate, ok := enumStringArrayValueCandidateForItemSchema(value, itemSchema); ok {
			return candidate, true
		}
	}
	return nil, false
}

func enumStringArrayValueCandidateForItemSchema(value string, itemSchema *jsonschema.Schema) ([]any, bool) {
	itemSchema = resolveSchemaRef(itemSchema)
	if itemSchema == nil || itemSchema.Enum == nil || !schemaAllowsStringValue(itemSchema) {
		return nil, false
	}
	if !enumValuesAreAllStrings(itemSchema.Enum.Values) {
		return nil, false
	}

	parts := strings.Split(value, ",")
	candidate := make([]any, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			return nil, false
		}

		matches := enumStringMatchesByNormalizedValue(itemSchema.Enum.Values, normalizeEnumString(token))
		if len(matches) != 1 {
			return nil, false
		}
		candidate = append(candidate, matches[0])
	}
	return candidate, true
}

func enumValuesAreAllStrings(values []any) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if _, ok := value.(string); !ok {
			return false
		}
	}
	return true
}
