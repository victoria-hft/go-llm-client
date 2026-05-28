package schema_compliance

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var numericDatePattern = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)

func repairScalarSchemaValues(current string, schema *jsonschema.Schema) (string, bool) {
	value, _, err := parseAndNormalizeJSON(current)
	if err != nil {
		return "", false
	}

	currentLoss := schemaLoss(current, schema)
	var repaired string
	found := enumerateScalarSchemaValueCandidates(value, schema, func(candidate any) bool {
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

func enumerateScalarSchemaValueCandidates(value any, schema *jsonschema.Schema, yield func(any) bool) bool {
	branches := candidateSchemaBranches(schema)

	if text, ok := value.(string); ok {
		for _, branch := range branches {
			if candidate, ok := scalarSchemaValueCandidate(text, branch); ok {
				if yield(candidate) {
					return true
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
				if enumerateScalarSchemaValueCandidates(child, propertySchema, func(candidateChild any) bool {
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
					if enumerateScalarSchemaValueCandidates(item, itemSchema, func(candidateItem any) bool {
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

func scalarSchemaValueCandidate(value string, schema *jsonschema.Schema) (any, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil, false
	}

	if schemaAllowsType(schema, "null") && isPlaceholderString(value) {
		return nil, true
	}
	if schemaExpectsISODate(schema) {
		if isoDate, ok := parseConservativeDate(value); ok {
			return isoDate, true
		}
	}
	if schemaAllowsType(schema, "integer") {
		if number, ok := parseIntegerString(value); ok {
			return number, true
		}
	}
	if schemaAllowsType(schema, "number") {
		if number, ok := parseNumberString(value); ok {
			return number, true
		}
	}

	return nil, false
}

func schemaExpectsISODate(schema *jsonschema.Schema) bool {
	return schema != nil &&
		schema.Format != nil &&
		schema.Format.Name == "date" &&
		(schema.Types == nil || schema.Types.IsEmpty() || schemaAllowsType(schema, "string"))
}

func parseConservativeDate(value string) (string, bool) {
	trimmed := normalizeDateInput(value)
	if trimmed == "" {
		return "", false
	}
	if isAmbiguousNumericDate(trimmed) {
		return "", false
	}

	layouts := []string{
		"2 Jan 2006",
		"2 January 2006",
		"Jan 2 2006",
		"January 2 2006",
		"2006/01/02",
		"2/1/2006",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed.Format(time.DateOnly), true
		}
	}
	return "", false
}

func normalizeDateInput(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	for i, field := range fields {
		fields[i] = strings.TrimSuffix(field, ",")
	}
	return strings.Join(fields, " ")
}

func isAmbiguousNumericDate(value string) bool {
	matches := numericDatePattern.FindStringSubmatch(value)
	if matches == nil {
		return false
	}
	first, err := strconv.Atoi(matches[1])
	if err != nil {
		return true
	}
	second, err := strconv.Atoi(matches[2])
	if err != nil {
		return true
	}
	return first <= 12 && second <= 12
}

func parseIntegerString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.ContainsAny(trimmed, ".eE") {
		return "", false
	}
	if _, err := strconv.ParseInt(trimmed, 10, 64); err != nil {
		return "", false
	}
	return json.Number(trimmed), true
}

func parseNumberString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
		return "", false
	}
	if !json.Valid([]byte(trimmed)) {
		return "", false
	}
	return json.Number(trimmed), true
}

func isPlaceholderString(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "n/a", "na", "none", "null", "nil", "unknown", "not available", "not applicable", "-", "--":
		return true
	default:
		return false
	}
}

func arrayItemSchemas(schema *jsonschema.Schema) []*jsonschema.Schema {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil
	}

	var schemas []*jsonschema.Schema
	if itemSchema, ok := schema.Items.(*jsonschema.Schema); ok && itemSchema != nil {
		schemas = append(schemas, itemSchema)
	}
	if schema.Items2020 != nil {
		schemas = append(schemas, schema.Items2020)
	}
	return schemas
}

func itemSchemasForIndex(itemSchemas []*jsonschema.Schema, schema *jsonschema.Schema, index int) []*jsonschema.Schema {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return itemSchemas
	}
	if index < len(schema.PrefixItems) && schema.PrefixItems[index] != nil {
		return append([]*jsonschema.Schema{schema.PrefixItems[index]}, itemSchemas...)
	}
	return itemSchemas
}
