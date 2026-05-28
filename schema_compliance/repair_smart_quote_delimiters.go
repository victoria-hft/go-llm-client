package schema_compliance

import (
	"strings"
	"unicode"
)

func repairSmartQuoteDelimiters(output string) (string, bool) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" || (trimmed[0] != '{' && trimmed[0] != '[') {
		return output, false
	}

	runes := []rune(output)
	var builder strings.Builder
	builder.Grow(len(output))

	inString := false
	smartString := false
	escaped := false
	changed := false
	previousSignificant := rune(0)

	for index, r := range runes {
		if inString {
			if escaped {
				escaped = false
				builder.WriteRune(r)
				continue
			}
			if r == '\\' {
				escaped = true
				builder.WriteRune(r)
				continue
			}
			if smartString {
				if isSmartCloseQuote(r) && nextSignificantIsStringTerminator(runes, index+1) {
					inString = false
					smartString = false
					previousSignificant = '"'
					builder.WriteByte('"')
					changed = true
					continue
				}
				builder.WriteRune(r)
				continue
			}
			if r == '"' {
				inString = false
				previousSignificant = r
			}
			builder.WriteRune(r)
			continue
		}

		if isSmartOpenQuote(r) && canStartJSONStringAfter(previousSignificant) {
			inString = true
			smartString = true
			builder.WriteByte('"')
			changed = true
			continue
		}
		if r == '"' {
			inString = true
			smartString = false
			builder.WriteRune(r)
			continue
		}
		if !unicode.IsSpace(r) {
			previousSignificant = r
		}
		builder.WriteRune(r)
	}

	if inString || !changed {
		return output, false
	}
	return builder.String(), true
}

func canStartJSONStringAfter(previous rune) bool {
	return previous == 0 || previous == '{' || previous == '[' || previous == ':' || previous == ','
}

func nextSignificantIsStringTerminator(runes []rune, start int) bool {
	for index := start; index < len(runes); index++ {
		r := runes[index]
		if unicode.IsSpace(r) {
			continue
		}
		return r == ':' || r == ',' || r == '}' || r == ']'
	}
	return true
}

func isSmartOpenQuote(r rune) bool {
	return r == '“'
}

func isSmartCloseQuote(r rune) bool {
	return r == '”'
}
