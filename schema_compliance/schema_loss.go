package schema_compliance

import (
	"encoding/json"
	"reflect"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	invalidJSONLoss          = 1_000_000
	validationFallbackLoss   = 1_000
	typeMismatchLoss         = 30
	missingRequiredFieldLoss = 40
	unexpectedFieldLoss      = 20
	wrapperContainerLoss     = 10_000
	arrayWrapperLoss         = 50
	objectArrayWrapperLoss   = 10_000
	enumOrConstMismatchLoss  = 10
	canonicalFormatLoss      = 5
)

func schemaLoss(jsonText string, schema *jsonschema.Schema) int {
	value, _, err := parseAndNormalizeJSON(jsonText)
	if err != nil {
		return invalidJSONLoss
	}
	if err := schema.Validate(value); err == nil {
		return canonicalFormatValueLoss(value, resolveSchemaRef(schema), "")
	}
	loss := valueLoss(value, resolveSchemaRef(schema))
	if loss <= 0 {
		return validationFallbackLoss
	}
	return loss
}

func canonicalFormatValueLoss(value any, schema *jsonschema.Schema, propertyName string) int {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return 0
	}

	loss := 0
	if text, ok := value.(string); ok && schemaExpectsISODateTime(schema) {
		if canonical, ok := parseConservativeDateTime(text); ok && canonical != text {
			loss += canonicalFormatLoss
		}
	}
	if text, ok := value.(string); ok && (schemaExpectsURL(schema) || propertyNameLooksLikeURL(propertyName)) {
		if canonical, ok := parseConservativeURL(text); ok && canonical != text {
			loss += canonicalFormatLoss
		}
	}

	if object, ok := value.(map[string]any); ok {
		for name, propertySchema := range schema.Properties {
			if propertyValue, ok := object[name]; ok {
				loss += canonicalFormatValueLoss(propertyValue, propertySchema, name)
			}
		}
	}

	if array, ok := value.([]any); ok {
		for index, item := range array {
			itemLoss := 0
			foundSchema := false
			for _, itemSchema := range itemSchemasForIndex(arrayItemSchemas(schema), schema, index) {
				lossForSchema := canonicalFormatValueLoss(item, itemSchema, propertyName)
				if !foundSchema || lossForSchema < itemLoss {
					itemLoss = lossForSchema
				}
				foundSchema = true
			}
			loss += itemLoss
		}
	}

	if len(schema.AllOf) > 0 {
		for _, branch := range schema.AllOf {
			loss += canonicalFormatValueLoss(value, branch, propertyName)
		}
	}
	if len(schema.AnyOf) > 0 {
		loss += closestCanonicalFormatBranchLoss(value, schema.AnyOf, propertyName)
	}
	if len(schema.OneOf) > 0 {
		loss += closestCanonicalFormatBranchLoss(value, schema.OneOf, propertyName)
	}
	return loss
}

func closestCanonicalFormatBranchLoss(value any, branches []*jsonschema.Schema, propertyName string) int {
	if len(branches) == 0 {
		return 0
	}

	best := validationFallbackLoss
	for _, branch := range branches {
		if loss := canonicalFormatValueLoss(value, branch, propertyName); loss < best {
			best = loss
		}
	}
	return best
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
	if schema.Format != nil && schema.Format.Validate(value) != nil {
		loss += enumOrConstMismatchLoss
	}

	if object, ok := value.(map[string]any); ok {
		loss += objectLoss(object, schema)
		if schemaExpectsArray(schema) {
			loss += objectArrayWrapperLoss
		}
	} else if schemaExpectsObject(schema) {
		loss += missingRequiredLoss(schema)
	}
	if array, ok := value.([]any); ok {
		loss += arrayLoss(array, schema)
		if !schemaAllowsType(schema, "array") {
			loss += arrayWrapperLoss
		}
	}

	loss += compositionLoss(value, schema)

	if hasUnsupportedSchemaFeatures(schema) {
		loss += validationFallbackLoss
	}
	return loss
}

func compositionLoss(value any, schema *jsonschema.Schema) int {
	loss := 0
	if len(schema.AllOf) > 0 {
		for _, branch := range schema.AllOf {
			loss += valueLoss(value, branch)
		}
	}
	if len(schema.AnyOf) > 0 {
		loss += closestBranchLoss(value, schema.AnyOf)
	}
	if len(schema.OneOf) > 0 {
		loss += closestBranchLoss(value, schema.OneOf)
	}
	return loss
}

func closestBranchLoss(value any, branches []*jsonschema.Schema) int {
	if len(branches) == 0 {
		return 0
	}

	best := validationFallbackLoss
	for _, branch := range branches {
		if loss := valueLoss(value, branch); loss < best {
			best = loss
		}
	}
	return best
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
	itemSchemas := arrayItemSchemas(schema)
	if len(itemSchemas) == 0 {
		return 0
	}

	loss := 0
	for index, item := range array {
		itemLoss := validationFallbackLoss
		for _, itemSchema := range itemSchemasForIndex(itemSchemas, schema, index) {
			if lossForSchema := valueLoss(item, itemSchema); lossForSchema < itemLoss {
				itemLoss = lossForSchema
			}
		}
		loss += itemLoss
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
	return schema.Not != nil ||
		schema.If != nil ||
		schema.Then != nil ||
		schema.Else != nil
}
