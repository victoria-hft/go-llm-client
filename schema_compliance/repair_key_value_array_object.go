package schema_compliance

import "github.com/santhosh-tekuri/jsonschema/v6"

var keyValueArrayObjectFieldPairs = [][2]string{
	{"key", "value"},
	{"name", "value"},
	{"field", "value"},
	{"property", "value"},
	{"id", "value"},
	{"key", "val"},
	{"name", "val"},
}

func repairKeyValueArrayObject(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateKeyValueArrayObjectCandidates(value, schema, func(candidate any) bool {
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

func enumerateKeyValueArrayObjectCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if array, ok := value.([]any); ok {
		for _, branch := range branches {
			if schemaExpectsObject(branch) {
				if candidate, ok := keyValueArrayObjectCandidate(array); ok {
					if yield(candidate) {
						return true
					}
				}
			}
		}

		for _, branch := range branches {
			for index, item := range array {
				for _, itemSchema := range itemSchemasForIndex(arrayItemSchemas(branch), branch, index) {
					if enumerateKeyValueArrayObjectCandidates(item, itemSchema, func(candidateItem any) bool {
						candidate := cloneJSONArray(array)
						candidate[index] = candidateItem
						return yield(candidate)
					}) {
						return true
					}
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
				if enumerateKeyValueArrayObjectCandidates(child, propertySchema, func(candidateChild any) bool {
					candidate := cloneJSONObject(object)
					candidate[key] = candidateChild
					return yield(candidate)
				}) {
					return true
				}
			}
		}
	}

	return false
}

func keyValueArrayObjectCandidate(array []any) (map[string]any, bool) {
	if len(array) == 0 {
		return nil, false
	}

	candidate := make(map[string]any, len(array))
	for _, item := range array {
		object, ok := item.(map[string]any)
		if !ok || len(object) != 2 {
			return nil, false
		}

		keyField, valueField, ok := keyValueFieldPairForObject(object)
		if !ok {
			return nil, false
		}
		key, ok := object[keyField].(string)
		if !ok {
			return nil, false
		}
		if _, exists := candidate[key]; exists {
			return nil, false
		}
		candidate[key] = cloneJSONValue(object[valueField])
	}
	return candidate, true
}

func keyValueFieldPairForObject(object map[string]any) (string, string, bool) {
	var matchedKey string
	var matchedValue string
	matches := 0
	for _, pair := range keyValueArrayObjectFieldPairs {
		keyField := pair[0]
		valueField := pair[1]
		if _, hasKey := object[keyField]; !hasKey {
			continue
		}
		if _, hasValue := object[valueField]; !hasValue {
			continue
		}
		matchedKey = keyField
		matchedValue = valueField
		matches++
	}
	return matchedKey, matchedValue, matches == 1
}
