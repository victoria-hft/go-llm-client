package schema_compliance

import "strings"

func repairTruncatedJSON(output string) (string, bool) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return output, false
	}

	closers, ok := missingJSONClosers(trimmed)
	if !ok || len(closers) == 0 {
		return output, false
	}
	candidate := trimmed + closers
	parser := newRelaxedJSONParser(candidate)
	if _, ok := parser.parse(); !ok {
		return output, false
	}
	return candidate, true
}

func missingJSONClosers(input string) (string, bool) {
	var stack []byte
	var quote byte
	escaped := false
	lastSignificant := byte(0)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
				lastSignificant = ch
			}
			continue
		}

		if isJSONWhitespace(ch) {
			continue
		}

		switch ch {
		case '"', '\'':
			quote = ch
		case '{':
			stack = append(stack, '}')
			lastSignificant = ch
		case '[':
			stack = append(stack, ']')
			lastSignificant = ch
		case '}', ']':
			if len(stack) == 0 || stack[len(stack)-1] != ch {
				return "", false
			}
			stack = stack[:len(stack)-1]
			lastSignificant = ch
		default:
			lastSignificant = ch
		}
	}

	if quote != 0 || escaped || len(stack) == 0 {
		return "", false
	}
	if !canCloseAfter(lastSignificant) {
		return "", false
	}

	var builder strings.Builder
	builder.Grow(len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		builder.WriteByte(stack[i])
	}
	return builder.String(), true
}

func canCloseAfter(ch byte) bool {
	if ch == '"' || ch == '\'' || ch == '}' || ch == ']' {
		return true
	}
	if ch == 'e' || ch == 'l' {
		return true
	}
	return ('0' <= ch && ch <= '9')
}

func isJSONWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t'
}
