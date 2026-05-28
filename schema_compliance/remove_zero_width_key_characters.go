package schema_compliance

import "strings"

func removeZeroWidthCharactersFromKeys(output string) (string, bool) {
	value, _, err := parseAndNormalizeJSON(output)
	if err != nil {
		return output, false
	}

	cleaned, changed, ok := cleanZeroWidthKeyCharacters(value)
	if !ok || !changed {
		return output, false
	}

	next, err := marshalCanonicalJSON(cleaned)
	if err != nil {
		return output, false
	}
	return next, true
}

func cleanZeroWidthKeyCharacters(value any) (any, bool, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return cleanZeroWidthObjectKeyCharacters(typed)
	case []any:
		return cleanZeroWidthArrayKeyCharacters(typed)
	default:
		return value, false, true
	}
}

func cleanZeroWidthObjectKeyCharacters(object map[string]any) (map[string]any, bool, bool) {
	cleaned := make(map[string]any, len(object))
	changed := false

	for _, key := range sortedObjectKeys(object) {
		value, childChanged, ok := cleanZeroWidthKeyCharacters(object[key])
		if !ok {
			return nil, false, false
		}

		cleanedKey := removeZeroWidthCharacters(key)
		if cleanedKey != key {
			changed = true
		}
		if childChanged {
			changed = true
		}
		if _, exists := cleaned[cleanedKey]; exists {
			return nil, false, false
		}
		cleaned[cleanedKey] = value
	}

	return cleaned, changed, true
}

func cleanZeroWidthArrayKeyCharacters(array []any) ([]any, bool, bool) {
	cleaned := make([]any, len(array))
	changed := false

	for index, item := range array {
		value, childChanged, ok := cleanZeroWidthKeyCharacters(item)
		if !ok {
			return nil, false, false
		}
		if childChanged {
			changed = true
		}
		cleaned[index] = value
	}

	return cleaned, changed, true
}

func removeZeroWidthCharacters(value string) string {
	if !strings.ContainsAny(value, "\u200b\u200c\u200d\u2060\ufeff") {
		return value
	}

	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		if isZeroWidthKeyCharacter(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func isZeroWidthKeyCharacter(r rune) bool {
	switch r {
	case '\u200b', '\u200c', '\u200d', '\u2060', '\ufeff':
		return true
	default:
		return false
	}
}
