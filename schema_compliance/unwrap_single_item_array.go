package schema_compliance

import "github.com/santhosh-tekuri/jsonschema/v6"

func unwrapSingleItemArray(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return current, false
	}

	array, ok := value.([]any)
	if !ok || len(array) != 1 {
		return current, false
	}

	candidate, err := marshalCanonicalJSON(array[0])
	if err != nil {
		return current, false
	}
	if schemaLoss(candidate, schema) >= schemaLoss(current, schema) {
		return current, false
	}
	return candidate, true
}
