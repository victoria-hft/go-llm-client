package schema_compliance

import (
	"sort"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repairObjectFieldNesting(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateFieldNestingCandidates(value, schema, func(candidate any) bool {
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

func enumerateFieldNestingCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if object, ok := value.(map[string]any); ok {
		for _, branch := range branches {
			if !schemaExpectsObject(branch) {
				continue
			}
			if enumeratePromotedSubfieldCandidates(object, branch, yield) {
				return true
			}
			if enumerateNestedSubfieldCandidates(object, branch, yield) {
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
				if enumerateFieldNestingCandidates(child, propertySchema, func(candidateChild any) bool {
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
			itemSchema, ok := branch.Items.(*jsonschema.Schema)
			if !ok || itemSchema == nil {
				continue
			}
			for index, item := range array {
				if enumerateFieldNestingCandidates(item, itemSchema, func(candidateItem any) bool {
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

func enumeratePromotedSubfieldCandidates(object map[string]any, schema *jsonschema.Schema, yield func(any) bool) bool {
	for _, propertyName := range sortedSchemaPropertyNames(schema) {
		if _, exists := object[propertyName]; exists {
			continue
		}

		propertySchema := resolveSchemaRef(schema.Properties[propertyName])
		if !schemaExpectsObject(propertySchema) {
			continue
		}

		moved := make(map[string]any)
		for _, childName := range sortedSchemaPropertyNames(propertySchema) {
			childValue, exists := object[childName]
			if !exists {
				continue
			}
			moved[childName] = cloneJSONValue(childValue)
		}
		if len(moved) == 0 {
			continue
		}

		candidate := cloneJSONObject(object)
		for childName := range moved {
			delete(candidate, childName)
		}
		candidate[propertyName] = moved
		if yield(candidate) {
			return true
		}
	}
	return false
}

func enumerateNestedSubfieldCandidates(object map[string]any, schema *jsonschema.Schema, yield func(any) bool) bool {
	for _, wrapperName := range sortedObjectKeys(object) {
		if _, known := schema.Properties[wrapperName]; known {
			continue
		}

		wrapper, ok := object[wrapperName].(map[string]any)
		if !ok {
			continue
		}

		moved := make(map[string]any)
		for _, childName := range sortedObjectKeys(wrapper) {
			if _, alreadyExists := object[childName]; alreadyExists {
				continue
			}
			if _, expected := schema.Properties[childName]; !expected {
				continue
			}
			moved[childName] = cloneJSONValue(wrapper[childName])
		}
		if len(moved) == 0 {
			continue
		}

		candidate := cloneJSONObject(object)
		candidateWrapper := cloneJSONObject(wrapper)
		for childName, childValue := range moved {
			delete(candidateWrapper, childName)
			candidate[childName] = childValue
		}
		if len(candidateWrapper) == 0 {
			delete(candidate, wrapperName)
		} else {
			candidate[wrapperName] = candidateWrapper
		}
		if yield(candidate) {
			return true
		}
	}
	return false
}

func candidateSchemaBranches(schema *jsonschema.Schema) []*jsonschema.Schema {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil
	}

	branches := []*jsonschema.Schema{schema}
	for _, branch := range schema.OneOf {
		branches = append(branches, candidateSchemaBranches(branch)...)
	}
	for _, branch := range schema.AnyOf {
		branches = append(branches, candidateSchemaBranches(branch)...)
	}
	for _, branch := range schema.AllOf {
		branches = append(branches, candidateSchemaBranches(branch)...)
	}
	return branches
}

func sortedSchemaPropertyNames(schema *jsonschema.Schema) []string {
	if schema == nil || len(schema.Properties) == 0 {
		return nil
	}
	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedObjectKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneJSONObject(typed)
	case []any:
		return cloneJSONArray(typed)
	default:
		return typed
	}
}

func cloneJSONObject(object map[string]any) map[string]any {
	cloned := make(map[string]any, len(object))
	for key, value := range object {
		cloned[key] = cloneJSONValue(value)
	}
	return cloned
}

func cloneJSONArray(array []any) []any {
	cloned := make([]any, len(array))
	for index, value := range array {
		cloned[index] = cloneJSONValue(value)
	}
	return cloned
}
