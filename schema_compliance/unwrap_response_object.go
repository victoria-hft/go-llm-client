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
	for key, value := range object {
		if _, ok := responseWrapperKeys[key]; !ok {
			return current, false
		}
		wrapped = value
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
