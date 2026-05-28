package schema_compliance

func stripTransportJunk(output string) (string, bool) {
	trimmed := output
	for trimmed != "" {
		if hasPrefixAny(trimmed, "\ufeff", "ï»¿", "���") {
			trimmed = trimKnownTransportJunkPrefix(trimmed)
			continue
		}

		if isJSONStartByte(trimmed[0]) {
			break
		}
		if !isTransportJunkByte(trimmed[0]) {
			break
		}
		trimmed = trimmed[1:]
	}

	if trimmed == "" || trimmed == output {
		return output, false
	}
	return trimmed, true
}

func trimKnownTransportJunkPrefix(input string) string {
	for _, prefix := range []string{"\ufeff", "ï»¿", "���"} {
		if len(input) >= len(prefix) && input[:len(prefix)] == prefix {
			return input[len(prefix):]
		}
	}
	return input
}

func hasPrefixAny(input string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if len(input) >= len(prefix) && input[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func isTransportJunkByte(ch byte) bool {
	return ch == 0x00 || ch == 0x1a || ch == 0xef || ch == 0xbb || ch == 0xbf
}

func isJSONStartByte(ch byte) bool {
	return ch == '{' || ch == '[' || ch == '"' || ch == '-' ||
		ch == 't' || ch == 'f' || ch == 'n' ||
		('0' <= ch && ch <= '9')
}
