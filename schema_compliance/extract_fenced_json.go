package schema_compliance

import "strings"

func extractSurroundedFencedJSON(output string) (string, bool) {
	start := strings.Index(output, "```")
	if start == -1 {
		return output, false
	}

	bodyStart := start + len("```")
	end := strings.Index(output[bodyStart:], "```")
	if end == -1 {
		return output, false
	}
	end += bodyStart

	body := strings.TrimSpace(output[bodyStart:end])
	if body == "" {
		return output, false
	}

	if startsWithJSONFenceLanguage(body) {
		withoutLanguage := strings.TrimSpace(body[len("json"):])
		if withoutLanguage != "" {
			body = withoutLanguage
		}
	}

	if body == "" {
		return output, false
	}
	return body, true
}

func startsWithJSONFenceLanguage(body string) bool {
	if len(body) < len("json") || !strings.EqualFold(body[:len("json")], "json") {
		return false
	}
	return len(body) == len("json") || isWhitespace(body[len("json")])
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}
