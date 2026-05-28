package schema_compliance

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

func repairRelaxedJSON(output string) (string, bool) {
	parser := newRelaxedJSONParser(output)
	value, ok := parser.parse()
	if !ok {
		return output, false
	}

	var repaired bytes.Buffer
	encoder := json.NewEncoder(&repaired)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return output, false
	}

	next := strings.TrimSpace(repaired.String())
	if next == strings.TrimSpace(output) {
		return output, false
	}
	return next, true
}

type relaxedJSONParser struct {
	input string
	pos   int
}

func newRelaxedJSONParser(input string) *relaxedJSONParser {
	return &relaxedJSONParser{input: strings.TrimSpace(input)}
}

func (p *relaxedJSONParser) parse() (any, bool) {
	value, ok := p.parseValue()
	if !ok {
		return nil, false
	}
	p.skipWhitespace()
	if p.pos != len(p.input) {
		return nil, false
	}
	return value, true
}

func (p *relaxedJSONParser) parseValue() (any, bool) {
	p.skipWhitespace()
	if p.done() {
		return nil, false
	}

	switch p.peek() {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		return p.parseDoubleQuotedString()
	case '\'':
		return p.parseSingleQuotedString()
	default:
		if p.peek() == '-' || isDigit(p.peek()) {
			if value, ok := p.parseHexIntegerLiteralString(); ok {
				return value, true
			}
			return p.parseNumber()
		}
		return p.parseBareValue()
	}
}

func (p *relaxedJSONParser) parseObject() (any, bool) {
	if !p.consume('{') {
		return nil, false
	}

	object := map[string]any{}
	p.skipWhitespace()
	if p.consume('}') {
		return object, true
	}

	for {
		key, ok := p.parseObjectKey()
		if !ok {
			return nil, false
		}

		p.skipWhitespace()
		if !p.consume(':') {
			return nil, false
		}

		value, ok := p.parseValue()
		if !ok {
			return nil, false
		}
		object[key] = value

		p.skipWhitespace()
		if p.consume('}') {
			return object, true
		}
		if !p.consume(',') {
			return nil, false
		}
		p.skipWhitespace()
		if p.consume('}') {
			return object, true
		}
	}
}

func (p *relaxedJSONParser) parseArray() (any, bool) {
	if !p.consume('[') {
		return nil, false
	}

	var array []any
	p.skipWhitespace()
	if p.consume(']') {
		return array, true
	}

	for {
		value, ok := p.parseValue()
		if !ok {
			return nil, false
		}
		array = append(array, value)

		p.skipWhitespace()
		if p.consume(']') {
			return array, true
		}
		if !p.consume(',') {
			return nil, false
		}
		p.skipWhitespace()
		if p.consume(']') {
			return array, true
		}
	}
}

func (p *relaxedJSONParser) parseObjectKey() (string, bool) {
	p.skipWhitespace()
	if p.done() {
		return "", false
	}

	switch p.peek() {
	case '"':
		return p.parseDoubleQuotedString()
	case '\'':
		return p.parseSingleQuotedString()
	default:
		return p.parseIdentifier()
	}
}

func (p *relaxedJSONParser) parseDoubleQuotedString() (string, bool) {
	start := p.pos
	if !p.consume('"') {
		return "", false
	}

	escaped := false
	for !p.done() {
		ch := p.next()
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			var value string
			if err := json.Unmarshal([]byte(p.input[start:p.pos]), &value); err != nil {
				return "", false
			}
			return value, true
		}
	}
	return "", false
}

func (p *relaxedJSONParser) parseSingleQuotedString() (string, bool) {
	if !p.consume('\'') {
		return "", false
	}

	var builder strings.Builder
	for !p.done() {
		ch := p.next()
		if ch == '\'' {
			return builder.String(), true
		}
		if ch == '\\' {
			if p.done() {
				return "", false
			}
			escaped := p.next()
			switch escaped {
			case '\\', '\'':
				builder.WriteByte(escaped)
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			default:
				return "", false
			}
			continue
		}
		builder.WriteByte(ch)
	}
	return "", false
}

func (p *relaxedJSONParser) parseNumber() (any, bool) {
	start := p.pos
	if p.peek() == '-' {
		p.pos++
	}
	if p.done() {
		return nil, false
	}
	if p.peek() == '0' {
		p.pos++
	} else if isDigitOneToNine(p.peek()) {
		for !p.done() && isDigit(p.peek()) {
			p.pos++
		}
	} else {
		return nil, false
	}
	if !p.done() && p.peek() == '.' {
		p.pos++
		if p.done() || !isDigit(p.peek()) {
			return nil, false
		}
		for !p.done() && isDigit(p.peek()) {
			p.pos++
		}
	}
	if !p.done() && (p.peek() == 'e' || p.peek() == 'E') {
		p.pos++
		if !p.done() && (p.peek() == '+' || p.peek() == '-') {
			p.pos++
		}
		if p.done() || !isDigit(p.peek()) {
			return nil, false
		}
		for !p.done() && isDigit(p.peek()) {
			p.pos++
		}
	}

	number := p.input[start:p.pos]
	if _, err := strconv.ParseFloat(number, 64); err != nil {
		return nil, false
	}
	return json.Number(number), true
}

func (p *relaxedJSONParser) parseHexIntegerLiteralString() (string, bool) {
	start := p.pos
	if !p.done() && p.peek() == '-' {
		p.pos++
	}
	if p.pos+2 > len(p.input) || p.input[p.pos] != '0' || (p.input[p.pos+1] != 'x' && p.input[p.pos+1] != 'X') {
		p.pos = start
		return "", false
	}
	p.pos += 2
	digitStart := p.pos
	for !p.done() && isHexDigit(p.peek()) {
		p.pos++
	}
	if p.pos == digitStart {
		p.pos = start
		return "", false
	}
	return p.input[start:p.pos], true
}

func (p *relaxedJSONParser) parseBareValue() (any, bool) {
	value, ok := p.parseIdentifier()
	if !ok {
		return nil, false
	}

	switch value {
	case "true":
		return true, true
	case "false":
		return false, true
	case "null":
		return nil, true
	default:
		return value, true
	}
}

func (p *relaxedJSONParser) parseIdentifier() (string, bool) {
	if p.done() || !isIdentifierStart(p.peek()) {
		return "", false
	}

	start := p.pos
	p.pos++
	for !p.done() && isIdentifierPart(p.peek()) {
		p.pos++
	}
	return p.input[start:p.pos], true
}

func (p *relaxedJSONParser) skipWhitespace() {
	for !p.done() && unicode.IsSpace(rune(p.peek())) {
		p.pos++
	}
}

func (p *relaxedJSONParser) consume(ch byte) bool {
	p.skipWhitespace()
	if p.done() || p.peek() != ch {
		return false
	}
	p.pos++
	return true
}

func (p *relaxedJSONParser) peek() byte {
	return p.input[p.pos]
}

func (p *relaxedJSONParser) next() byte {
	ch := p.input[p.pos]
	p.pos++
	return ch
}

func (p *relaxedJSONParser) done() bool {
	return p.pos >= len(p.input)
}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || ch == '$' || ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func isIdentifierPart(ch byte) bool {
	return isIdentifierStart(ch) || isDigit(ch)
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isDigitOneToNine(ch byte) bool {
	return '1' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}
