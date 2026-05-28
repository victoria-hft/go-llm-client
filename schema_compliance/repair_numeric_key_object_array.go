package schema_compliance

import (
	"sort"
	"strconv"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repairNumericKeyObjectArray(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateNumericKeyObjectArrayCandidates(value, schema, func(candidate any) bool {
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

func enumerateNumericKeyObjectArrayCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if object, ok := value.(map[string]any); ok {
		for _, branch := range branches {
			if schemaExpectsArray(branch) {
				if candidate, ok := numericKeyObjectArrayCandidate(object); ok {
					if yield(candidate) {
						return true
					}
				}
			}
		}

		for _, key := range sortedObjectKeys(object) {
			child := object[key]
			for _, branch := range branches {
				propertySchema, ok := branch.Properties[key]
				if !ok {
					continue
				}
				if enumerateNumericKeyObjectArrayCandidates(child, propertySchema, func(candidateChild any) bool {
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
					if enumerateNumericKeyObjectArrayCandidates(item, itemSchema, func(candidateItem any) bool {
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

func numericKeyObjectArrayCandidate(object map[string]any) ([]any, bool) {
	if len(object) == 0 {
		return nil, false
	}

	byIndex := make(map[int]any, len(object))
	var indexes []int
	for key, value := range object {
		index, ok := parseArrayObjectKey(key)
		if !ok {
			return nil, false
		}
		if _, exists := byIndex[index]; exists {
			return nil, false
		}
		byIndex[index] = value
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	start := indexes[0]
	if start != 0 && start != 1 {
		return nil, false
	}
	for offset, index := range indexes {
		if index != start+offset {
			return nil, false
		}
	}

	candidate := make([]any, len(indexes))
	for offset, index := range indexes {
		candidate[offset] = cloneJSONValue(byIndex[index])
	}
	return candidate, true
}

func parseArrayObjectKey(key string) (int, bool) {
	if key == "" || key[0] == '-' || (len(key) > 1 && key[0] == '0') {
		return 0, false
	}
	index, err := strconv.Atoi(key)
	if err != nil {
		return 0, false
	}
	return index, true
}

func schemaExpectsArray(schema *jsonschema.Schema) bool {
	schema = resolveSchemaRef(schema)
	return schemaAllowsType(schema, "array") ||
		(schema != nil && (schema.Items != nil || schema.Items2020 != nil || len(schema.PrefixItems) > 0))
}
