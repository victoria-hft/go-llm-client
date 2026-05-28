package schema_compliance

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var numericDatePattern = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
var simpleDurationPattern = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s+([A-Za-z]+)$`)
var commaSeparatedNumberPattern = regexp.MustCompile(`^[+-]?\d{1,3}(,\d{3})+(\.\d+)?$`)
var underscoreSeparatedNumberPattern = regexp.MustCompile(`^[+-]?\d{1,3}(_\d{3})+(\.\d+)?$`)
var spaceSeparatedNumberPattern = regexp.MustCompile(`^[+-]?\d{1,3}( \d{3})+(\.\d+)?$`)
var hexIntegerPattern = regexp.MustCompile(`^[+-]?0[xX][0-9a-fA-F]+$`)
var bigIntPattern = regexp.MustCompile(`^[+-]?\d+n$`)
var fractionPattern = regexp.MustCompile(`^[+-]?\d+/[+-]?\d+$`)
var percentPattern = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?%$`)
var basisPointsPattern = regexp.MustCompile(`^([+-]?\d+(?:\.\d+)?)\s*(?i:bps)$`)

const (
	maxEpochSeconds = 32503680000
	maxEpochMillis  = maxEpochSeconds * 1000
	maxEpochMicros  = maxEpochSeconds * 1000 * 1000
)

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

	if number, ok := value.(json.Number); ok {
		for _, branch := range branches {
			if candidate, ok := scalarSchemaNumberCandidate(number, branch); ok {
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
	if schemaAllowsType(schema, "boolean") {
		if booleanValue, ok := parseBooleanMarkerString(value); ok {
			return booleanValue, true
		}
	}
	if schemaExpectsISODate(schema) {
		if isoDate, ok := parseEpochDate(value); ok {
			return isoDate, true
		}
		if isoDate, ok := parseConservativeDate(value); ok {
			return isoDate, true
		}
	}
	if schemaExpectsISODateTime(schema) {
		if isoDateTime, ok := parseEpochDateTime(value); ok {
			return isoDateTime, true
		}
		if isoDateTime, ok := parseConservativeDateTime(value); ok {
			return isoDateTime, true
		}
	}
	if schemaExpectsISODuration(schema) {
		if duration, ok := parseConservativeDuration(value); ok {
			return duration, true
		}
	}
	if schemaAllowsType(schema, "integer") {
		if number, ok := parseHexIntegerString(value); ok {
			return number, true
		}
		if number, ok := parseBigIntString(value); ok {
			return number, true
		}
		if number, ok := parseSeparatedIntegerString(value); ok {
			return number, true
		}
		if number, ok := parseIntegerString(value); ok {
			return number, true
		}
	}
	if schemaAllowsType(schema, "number") {
		if number, ok := parseSeparatedNumberString(value); ok {
			return number, true
		}
		if number, ok := parseFractionNumberString(value); ok {
			return number, true
		}
		if number, ok := parsePercentString(value, schema); ok {
			return number, true
		}
		if number, ok := parseBasisPointsString(value, schema); ok {
			return number, true
		}
		if number, ok := parseNumberString(value); ok {
			return number, true
		}
	}

	return nil, false
}

func scalarSchemaNumberCandidate(value json.Number, schema *jsonschema.Schema) (any, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil, false
	}

	if schemaExpectsISODate(schema) {
		if isoDate, ok := parseEpochDate(value.String()); ok {
			return isoDate, true
		}
	}
	if schemaExpectsISODateTime(schema) {
		if isoDateTime, ok := parseEpochDateTime(value.String()); ok {
			return isoDateTime, true
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

func schemaExpectsISODateTime(schema *jsonschema.Schema) bool {
	return schema != nil &&
		schema.Format != nil &&
		schema.Format.Name == "date-time" &&
		(schema.Types == nil || schema.Types.IsEmpty() || schemaAllowsType(schema, "string"))
}

func schemaExpectsISODuration(schema *jsonschema.Schema) bool {
	return schema != nil &&
		schema.Format != nil &&
		schema.Format.Name == "duration" &&
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

func parseConservativeDateTime(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < len("2006-01-02T15:04:05Z") {
		return "", false
	}

	candidate := trimmed
	if len(candidate) > 10 && candidate[10] == ' ' {
		candidate = candidate[:10] + "T" + candidate[11:]
	}
	if strings.HasSuffix(candidate, "z") {
		candidate = strings.TrimSuffix(candidate, "z") + "Z"
	}

	parsed, err := time.Parse(time.RFC3339Nano, candidate)
	if err != nil {
		return "", false
	}
	return parsed.UTC().Format(time.RFC3339Nano), true
}

func parseEpochDate(value string) (string, bool) {
	epochTime, ok := parseEpochTimestamp(value)
	if !ok {
		return "", false
	}
	return epochTime.Format(time.DateOnly), true
}

func parseEpochDateTime(value string) (string, bool) {
	epochTime, ok := parseEpochTimestamp(value)
	if !ok {
		return "", false
	}
	return epochTime.Format(time.RFC3339Nano), true
}

func parseEpochTimestamp(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			return time.Time{}, false
		}
	}
	if len(trimmed) > 1 && trimmed[0] == '0' {
		return time.Time{}, false
	}

	raw, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || raw < 0 || raw > maxEpochMicros {
		return time.Time{}, false
	}

	var epochTime time.Time
	switch {
	case raw <= maxEpochSeconds:
		epochTime = time.Unix(raw, 0)
	case raw <= maxEpochMillis:
		epochTime = time.Unix(raw/1000, (raw%1000)*int64(time.Millisecond))
	default:
		epochTime = time.Unix(raw/1000000, (raw%1000000)*int64(time.Microsecond))
	}
	epochTime = epochTime.UTC()
	if epochTime.Year() < 1970 || epochTime.Year() > 3000 {
		return time.Time{}, false
	}
	return epochTime, true
}

func parseConservativeDuration(value string) (string, bool) {
	matches := simpleDurationPattern.FindStringSubmatch(strings.TrimSpace(value))
	if matches == nil {
		return "", false
	}

	amount := matches[1]
	unit := strings.ToLower(matches[2])
	hasDecimal := strings.Contains(amount, ".")
	if hasDecimal {
		return "", false
	}
	if _, err := strconv.ParseUint(amount, 10, 64); err != nil {
		return "", false
	}

	switch unit {
	case "second", "seconds":
		return "PT" + amount + "S", true
	case "minute", "minutes":
		return "PT" + amount + "M", true
	case "hour", "hours":
		return "PT" + amount + "H", true
	case "day", "days":
		return "P" + amount + "D", true
	case "week", "weeks":
		return "P" + amount + "W", true
	default:
		return "", false
	}
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

func parseSeparatedIntegerString(value string) (json.Number, bool) {
	number, ok := parseSeparatedNumberString(value)
	if !ok || strings.Contains(number.String(), ".") {
		return "", false
	}
	if _, err := strconv.ParseInt(number.String(), 10, 64); err != nil {
		return "", false
	}
	return number, true
}

func parseSeparatedNumberString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	var cleaned string
	switch {
	case commaSeparatedNumberPattern.MatchString(trimmed):
		cleaned = strings.ReplaceAll(trimmed, ",", "")
	case underscoreSeparatedNumberPattern.MatchString(trimmed):
		cleaned = strings.ReplaceAll(trimmed, "_", "")
	case spaceSeparatedNumberPattern.MatchString(trimmed):
		cleaned = strings.ReplaceAll(trimmed, " ", "")
	default:
		return "", false
	}
	if !json.Valid([]byte(cleaned)) {
		return "", false
	}
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return "", false
	}
	return json.Number(cleaned), true
}

func parseHexIntegerString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if !hexIntegerPattern.MatchString(trimmed) {
		return "", false
	}

	sign := ""
	digits := trimmed
	if strings.HasPrefix(digits, "+") || strings.HasPrefix(digits, "-") {
		sign = digits[:1]
		digits = digits[1:]
	}

	parsed, ok := new(big.Int).SetString(digits[2:], 16)
	if !ok {
		return "", false
	}
	if sign == "-" {
		parsed.Neg(parsed)
	}
	if !parsed.IsInt64() {
		return "", false
	}
	return json.Number(strconv.FormatInt(parsed.Int64(), 10)), true
}

func parseBigIntString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if !bigIntPattern.MatchString(trimmed) {
		return "", false
	}
	digits := strings.TrimSuffix(trimmed, "n")
	parsed, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return "", false
	}
	return json.Number(strconv.FormatInt(parsed, 10)), true
}

func parseFractionNumberString(value string) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if !fractionPattern.MatchString(trimmed) {
		return "", false
	}

	parts := strings.Split(trimmed, "/")
	numerator, ok := new(big.Rat).SetString(parts[0])
	if !ok {
		return "", false
	}
	denominator, ok := new(big.Rat).SetString(parts[1])
	if !ok || denominator.Sign() == 0 {
		return "", false
	}

	result := new(big.Rat).Quo(numerator, denominator)
	return json.Number(ratToJSONNumber(result)), true
}

func parsePercentString(value string, schema *jsonschema.Schema) (json.Number, bool) {
	trimmed := strings.TrimSpace(value)
	if !percentPattern.MatchString(trimmed) {
		return "", false
	}
	numberText := strings.TrimSuffix(trimmed, "%")
	scale, ok := rateScaleForSchema(schema)
	if !ok {
		return "", false
	}

	valueRat, ok := new(big.Rat).SetString(numberText)
	if !ok {
		return "", false
	}
	result := new(big.Rat).Quo(valueRat, scale)
	return json.Number(ratToJSONNumber(result)), true
}

func parseBasisPointsString(value string, schema *jsonschema.Schema) (json.Number, bool) {
	matches := basisPointsPattern.FindStringSubmatch(strings.TrimSpace(value))
	if matches == nil {
		return "", false
	}
	scale, ok := rateScaleForSchema(schema)
	if !ok {
		return "", false
	}

	valueRat, ok := new(big.Rat).SetString(matches[1])
	if !ok {
		return "", false
	}
	result := new(big.Rat).Quo(valueRat, new(big.Rat).Mul(scale, big.NewRat(100, 1)))
	return json.Number(ratToJSONNumber(result)), true
}

func rateScaleForSchema(schema *jsonschema.Schema) (*big.Rat, bool) {
	schema = resolveSchemaRef(schema)
	if schema == nil {
		return nil, false
	}

	if schemaMinimumAllows(schema, 0) && schemaMaximumAtMost(schema, 1) {
		return big.NewRat(100, 1), true
	}
	if schemaMinimumAllows(schema, 0) && schemaMaximumGreaterThan(schema, 1) && schemaMaximumAtMost(schema, 100) {
		return big.NewRat(1, 1), true
	}
	return nil, false
}

func schemaMinimumAllows(schema *jsonschema.Schema, value int64) bool {
	target := big.NewRat(value, 1)
	if schema.Minimum != nil {
		return schema.Minimum.Cmp(target) <= 0
	}
	if schema.ExclusiveMinimum != nil {
		return schema.ExclusiveMinimum.Cmp(target) < 0
	}
	return false
}

func schemaMaximumAllows(schema *jsonschema.Schema, value int64) bool {
	target := big.NewRat(value, 1)
	if schema.Maximum != nil {
		return schema.Maximum.Cmp(target) >= 0
	}
	if schema.ExclusiveMaximum != nil {
		return schema.ExclusiveMaximum.Cmp(target) > 0
	}
	return false
}

func schemaMaximumAtMost(schema *jsonschema.Schema, value int64) bool {
	target := big.NewRat(value, 1)
	if schema.Maximum != nil {
		return schema.Maximum.Cmp(target) <= 0
	}
	if schema.ExclusiveMaximum != nil {
		return schema.ExclusiveMaximum.Cmp(target) <= 0
	}
	return false
}

func schemaMaximumGreaterThan(schema *jsonschema.Schema, value int64) bool {
	target := big.NewRat(value, 1)
	if schema.Maximum != nil {
		return schema.Maximum.Cmp(target) > 0
	}
	if schema.ExclusiveMaximum != nil {
		return schema.ExclusiveMaximum.Cmp(target) > 0
	}
	return false
}

func ratToJSONNumber(value *big.Rat) string {
	if value.IsInt() {
		return value.Num().String()
	}

	output := value.FloatString(16)
	output = strings.TrimRight(output, "0")
	output = strings.TrimRight(output, ".")
	if output == "-0" {
		return "0"
	}
	return output
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

func parseBooleanMarkerString(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "y", "✓", "✔", "☑", "✅":
		return true, true
	case "false", "no", "n", "✗", "✘", "✕", "×", "☒", "❌":
		return false, true
	default:
		return false, false
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
