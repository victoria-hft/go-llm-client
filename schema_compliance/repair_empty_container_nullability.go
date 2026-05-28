package schema_compliance

import "github.com/santhosh-tekuri/jsonschema/v6"

func repairEmptyContainerNullability(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateEmptyContainerNullabilityCandidates(value, schema, func(candidate any) bool {
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

func enumerateEmptyContainerNullabilityCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	for _, branch := range branches {
		if candidate, ok := localEmptyContainerNullabilityCandidate(value, branch); ok {
			if yield(candidate) {
				return true
			}
		}
	}

	if object, ok := value.(map[string]any); ok {
		for _, key := range sortedObjectKeys(object) {
			child := object[key]
			for _, branch := range branches {
				propertySchema, ok := branch.Properties[key]
				if !ok {
					continue
				}
				if enumerateEmptyContainerNullabilityCandidates(child, propertySchema, func(candidateChild any) bool {
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
					if enumerateEmptyContainerNullabilityCandidates(item, itemSchema, func(candidateItem any) bool {
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

func localEmptyContainerNullabilityCandidate(value any, schema *jsonschema.Schema) (any, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil, false
	}

	if object, ok := value.(map[string]any); ok && len(object) == 0 {
		if schemaValueValid(nil, schema) && !schemaValueValid(object, schema) {
			return nil, true
		}
		return nil, false
	}

	if array, ok := value.([]any); ok && len(array) == 0 {
		if schemaValueValid(nil, schema) && !schemaValueValid(array, schema) {
			return nil, true
		}
		return nil, false
	}

	if value == nil {
		emptyArray := []any{}
		if !schemaValueValid(nil, schema) && schemaValueValid(emptyArray, schema) {
			return emptyArray, true
		}
	}

	return nil, false
}

func schemaValueValid(value any, schema *jsonschema.Schema) bool {
	schema = resolveSchemaRef(schema)
	return schema != nil && schema.Validate(value) == nil
}
