package schema_compliance

import (
	"encoding/json"
	"reflect"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	invalidJSONLoss          = 1_000_000
	validationFallbackLoss   = 1_000
	typeMismatchLoss         = 100
	missingRequiredFieldLoss = 40
	unexpectedFieldLoss      = 20
	wrapperContainerLoss     = 100
	arrayWrapperLoss         = 50
	enumOrConstMismatchLoss  = 10
)

func schemaLoss(jsonText string, schema *jsonschema.Schema) int {
	value, _, err := parseAndNormalizeJSON(jsonText)
	if err != nil {
		return invalidJSONLoss
	}
	if err := schema.Validate(value); err == nil {
		return 0
	}
	loss := valueLoss(value, resolveSchemaRef(schema))
	if loss <= 0 {
		return validationFallbackLoss
	}
	return loss
}

func valueLoss(value any, schema *jsonschema.Schema) int {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return validationFallbackLoss
	}

	loss := 0
	if !schemaTypesMatch(value, schema) {
		loss += typeMismatchLoss
	}
	if schema.Const != nil && !reflect.DeepEqual(value, *schema.Const) {
		loss += enumOrConstMismatchLoss
	}
	if schema.Enum != nil && !enumContains(schema.Enum.Values, value) {
		loss += enumOrConstMismatchLoss
	}

	if object, ok := value.(map[string]any); ok {
		loss += objectLoss(object, schema)
	} else if schemaExpectsObject(schema) {
		loss += missingRequiredLoss(schema)
	}
	if array, ok := value.([]any); ok {
		loss += arrayLoss(array, schema)
		if !schemaAllowsType(schema, "array") {
			loss += arrayWrapperLoss
		}
	}

	if hasUnsupportedSchemaFeatures(schema) {
		loss += validationFallbackLoss
	}
	return loss
}

func objectLoss(object map[string]any, schema *jsonschema.Schema) int {
	loss := missingRequiredLossForObject(object, schema)
	for name, propertySchema := range schema.Properties {
		if propertyValue, ok := object[name]; ok {
			loss += valueLoss(propertyValue, propertySchema)
		}
	}
	if additional, ok := schema.AdditionalProperties.(bool); ok && !additional {
		for name, value := range object {
			if _, known := schema.Properties[name]; !known {
				loss += unexpectedFieldLoss
				if isContainer(value) && schemaExpectsObject(schema) {
					loss += wrapperContainerLoss
				}
			}
		}
	}
	return loss
}

func missingRequiredLossForObject(object map[string]any, schema *jsonschema.Schema) int {
	loss := 0
	for _, required := range schema.Required {
		if _, ok := object[required]; !ok {
			loss += missingRequiredFieldLoss
		}
	}
	return loss
}

func missingRequiredLoss(schema *jsonschema.Schema) int {
	return len(schema.Required) * missingRequiredFieldLoss
}

func arrayLoss(array []any, schema *jsonschema.Schema) int {
	itemSchema, ok := schema.Items.(*jsonschema.Schema)
	if !ok || itemSchema == nil {
		return 0
	}

	loss := 0
	for _, item := range array {
		loss += valueLoss(item, itemSchema)
	}
	return loss
}

func resolveSchemaRef(schema *jsonschema.Schema) *jsonschema.Schema {
	for schema != nil && schema.Ref != nil {
		schema = schema.Ref
	}
	return schema
}

func schemaTypesMatch(value any, schema *jsonschema.Schema) bool {
	if schema.Types == nil || schema.Types.IsEmpty() {
		return true
	}
	for _, typ := range schema.Types.ToStrings() {
		if valueMatchesType(value, typ) {
			return true
		}
	}
	return false
}

func schemaExpectsObject(schema *jsonschema.Schema) bool {
	return schemaAllowsType(schema, "object") ||
		len(schema.Required) > 0 ||
		len(schema.Properties) > 0 ||
		schema.AdditionalProperties != nil
}

func schemaAllowsType(schema *jsonschema.Schema, typ string) bool {
	if schema.Types == nil || schema.Types.IsEmpty() {
		return false
	}
	for _, allowed := range schema.Types.ToStrings() {
		if allowed == typ {
			return true
		}
	}
	return false
}

func valueMatchesType(value any, typ string) bool {
	switch typ {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		if ok {
			return true
		}
		_, ok = value.(float64)
		return ok
	case "integer":
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := number.Int64()
		return err == nil
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return true
	}
}

func enumContains(values []any, value any) bool {
	for _, enumValue := range values {
		if reflect.DeepEqual(enumValue, value) {
			return true
		}
	}
	return false
}

func isContainer(value any) bool {
	switch value.(type) {
	case map[string]any, []any:
		return true
	default:
		return false
	}
}

func hasUnsupportedSchemaFeatures(schema *jsonschema.Schema) bool {
	return len(schema.AllOf) > 0 ||
		len(schema.AnyOf) > 0 ||
		len(schema.OneOf) > 0 ||
		schema.Not != nil ||
		schema.If != nil ||
		schema.Then != nil ||
		schema.Else != nil
}
