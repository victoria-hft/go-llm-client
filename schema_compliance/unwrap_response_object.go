package schema_compliance

import "github.com/santhosh-tekuri/jsonschema/v6"

var responseWrapperKeys = map[string]struct{}{
	"data":     {},
	"result":   {},
	"results":  {},
	"response": {},
	"output":   {},
	"answer":   {},
	"content":  {},
	"payload":  {},
	"item":     {},
	"items":    {},
}

func unwrapResponseObject(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return current, false
	}

	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return current, false
	}

	var wrapped any
	var wrapperKey string
	for key, value := range object {
		if _, ok := responseWrapperKeys[key]; !ok {
			return current, false
		}
		wrapperKey = key
		wrapped = value
	}
	if shouldPreserveItemItemsWrapper(wrapperKey, schema) {
		return current, false
	}

	candidate, err := marshalCanonicalJSON(wrapped)
	if err != nil {
		return current, false
	}
	if schemaLoss(candidate, schema) >= schemaLoss(current, schema) {
		return current, false
	}
	return candidate, true
}

func shouldPreserveItemItemsWrapper(key string, schema *jsonschema.Schema) bool {
	for _, branch := range candidateSchemaBranches(schema) {
		switch key {
		case "item":
			if itemsSchema, ok := branch.Properties["items"]; ok && schemaExpectsArray(itemsSchema) {
				return true
			}
		case "items":
			if _, ok := branch.Properties["item"]; ok {
				return true
			}
		}
	}
	return false
}
