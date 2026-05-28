package schema_compliance

import "github.com/santhosh-tekuri/jsonschema/v6"

func repairItemItemsShape(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateItemItemsShapeCandidates(value, schema, func(candidate any) bool {
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

func enumerateItemItemsShapeCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if object, ok := value.(map[string]any); ok {
		for _, branch := range branches {
			if !schemaExpectsObject(branch) {
				continue
			}
			if enumerateLocalItemItemsShapeCandidates(object, branch, yield) {
				return true
			}
		}

		for _, key := range sortedObjectKeys(object) {
			child := object[key]
			for _, branch := range branches {
				propertySchema, ok := branch.Properties[key]
				if !ok {
					continue
				}
				if enumerateItemItemsShapeCandidates(child, propertySchema, func(candidateChild any) bool {
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
					if enumerateItemItemsShapeCandidates(item, itemSchema, func(candidateItem any) bool {
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

func enumerateLocalItemItemsShapeCandidates(object map[string]any, schema *jsonschema.Schema, yield func(any) bool) bool {
	if _, hasItems := object["items"]; !hasItems {
		if itemValue, hasItem := object["item"]; hasItem {
			if itemsSchema, definesItems := schema.Properties["items"]; definesItems && schemaExpectsArray(itemsSchema) {
				candidate := cloneJSONObject(object)
				delete(candidate, "item")
				if itemValue == nil {
					candidate["items"] = []any{}
				} else {
					candidate["items"] = []any{cloneJSONValue(itemValue)}
				}
				if yield(candidate) {
					return true
				}
			}
		}
	}

	if _, hasItem := object["item"]; hasItem {
		return false
	}

	itemsValue, hasItems := object["items"]
	if !hasItems {
		return false
	}
	itemSchema, definesItem := schema.Properties["item"]
	if !definesItem {
		return false
	}

	switch typed := itemsValue.(type) {
	case []any:
		switch len(typed) {
		case 0:
			if !schemaAllowsNull(itemSchema) {
				return false
			}
			candidate := cloneJSONObject(object)
			delete(candidate, "items")
			candidate["item"] = nil
			return yield(candidate)
		case 1:
			candidate := cloneJSONObject(object)
			delete(candidate, "items")
			candidate["item"] = cloneJSONValue(typed[0])
			return yield(candidate)
		default:
			return false
		}
	default:
		candidate := cloneJSONObject(object)
		delete(candidate, "items")
		candidate["item"] = cloneJSONValue(itemsValue)
		return yield(candidate)
	}
}

func schemaAllowsNull(schema *jsonschema.Schema) bool {
	for _, branch := range candidateSchemaBranches(schema) {
		if schemaAllowsType(branch, "null") {
			return true
		}
	}
	return false
}
