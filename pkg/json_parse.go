package lisp

import (
	"fmt"
	"golisp/parsing"
	"strconv"
	"strings"
)

func ParseJSONString(s string) any        { return ParseJSON(parsing.NewStringSource(s)) }
func ParseJSON(cs parsing.CharSource) any { return NewJSONParser(cs).parseJson() }

type JSONParser struct{ *parsing.BaseParser }

func NewJSONParser(cs parsing.CharSource) JSONParser {
	return JSONParser{BaseParser: parsing.NewBaseParser(cs)}
}

// General
func (j JSONParser) parseJson() any {
	if result := j.parseElement(); j.Eof() {
		return result
	}
	panic("end of JSON expected")
}

func (j JSONParser) parseElement() any {
	j.skipWhitespaces()
	result := j.parseValue()
	j.skipWhitespaces()
	return result

}

func (j JSONParser) parseValue() any {
	if j.Take('{') {
		return j.parseObject()
	} else if j.Take('[') {
		return j.parseArray()
	} else if j.Take('"') {
		return j.parseString()
	} else if j.Take('t') {
		j.ExpectString("rue")
		return true
	} else if j.Take('f') {
		j.ExpectString("alse")
		return false
	} else if j.Take('n') {
		j.ExpectString("ull")
		return nil
	} else {
		return j.parseNumber()
	}
}

// Arrays
func (j JSONParser) parseArray() []any {
	j.skipWhitespaces()
	if j.Take(']') {
		return []any{}
	}

	array := []any{j.parseElement()}
	for j.Take(',') {
		array = append(array, j.parseElement())
	}

	j.Expect(']')
	return array
}

// Maps
func (j JSONParser) parseObject() map[string]any {
	j.skipWhitespaces()
	if j.Take('}') {
		return map[string]any{}
	}
	obj := make(map[string]any)

	j.addMember(obj)
	for j.Take(',') {
		j.addMember(obj)
	}

	j.Expect('}')
	return obj
}

func (j JSONParser) addMember(obj map[string]any) {
	j.skipWhitespaces()
	j.Expect('"')
	key := j.parseString()
	j.skipWhitespaces()
	j.Expect(':')
	val := j.parseElement()
	obj[key] = val
}

// Strings
func (j JSONParser) parseString() string { // parse "([String]")
	var sb strings.Builder
	for !j.Take('"') {
		if j.Eof() {
			panic("string unterminated")
		}

		if j.Take('\\') {

			if j.escaped(&sb, '"', '"') ||
				j.escaped(&sb, '\\', '\\') ||
				j.escaped(&sb, '/', '/') ||
				j.escaped(&sb, '\b', 'b') ||
				j.escaped(&sb, '\f', 'f') ||
				j.escaped(&sb, '\n', 'n') ||
				j.escaped(&sb, '\r', 'r') ||
				j.escaped(&sb, '\t', 't') {
				// next char
			} else if j.Take('u') {
				value := 0
				for i := 0; i < 4; i++ {
					value <<= 4
					if j.Between('0', '9') {
						value = j.nextHex(value, '0')
					} else if j.Between('a', 'f') {
						value = j.nextHex(value, 'a'-10)
					} else if j.Between('A', 'F') {
						value = j.nextHex(value, 'A'-10)
					} else {
						panic("expected hex digit")
					}
				}
				sb.WriteRune(rune(value))
			} else {
				panic("Unknown escape character \\" + string(j.TakeNext()))
			}

		} else {
			sb.WriteRune(j.TakeNext())
		}
	}
	return sb.String()
}

func (j JSONParser) nextHex(value, delta int) int {
	value += int(j.TakeNext()) - delta
	return value
}

func (j JSONParser) escaped(sb *strings.Builder, character, expected rune) bool {
	var consumed = j.Take(expected)
	if consumed {
		sb.WriteRune(character)
	}
	return consumed
}

// Numbers
func (j JSONParser) parseNumber() float64 {
	var sb strings.Builder
	j.takeInteger(&sb)

	if j.Take('.') { // fraction
		sb.WriteRune('.')
		j.takeDigits(&sb)
	}

	if j.Take('e') || j.Take('E') { // exponent
		sb.WriteRune('e')
		if j.Take('+') { // sign of exponent
			// nothing
		} else if j.Take('-') {
			sb.WriteRune('-')
		}
		j.takeDigits(&sb)
	}

	val, err := strconv.ParseFloat(sb.String(), 64)
	if err != nil {
		panic("invalid number: " + err.Error())
	}
	return val
}

func (j JSONParser) takeDigits(sb *strings.Builder) {
	for j.Between('0', '9') {
		sb.WriteRune(j.TakeNext())
	}
}

func (j JSONParser) takeInteger(sb *strings.Builder) {
	if j.Take('-') {
		sb.WriteRune('-')
	}
	if j.Take('0') {
		sb.WriteRune('0')
	} else if j.Between('1', '9') {
		j.takeDigits(sb)
	} else {
		panic("invalid number")
	}
}

// Whitespaces
func (j JSONParser) skipWhitespaces() {
	for j.Take(' ') || j.Take('\t') || j.Take('\n') || j.Take('\r') {
	}
}

// not optimal
var _ = jsonToString

func jsonToString(obj any) string {
	switch t := obj.(type) {
	case string:
		return strconv.Quote(t)
	case float64:
		return strconv.FormatFloat(t, 'f', 12, 64)
	case bool:
		return strconv.FormatBool(t)
	case map[string]any:
		var sb strings.Builder
		isFirst := true
		sb.WriteRune('{')
		for k, v := range t {
			if !isFirst {
				sb.WriteRune(',')
			}
			isFirst = false
			sb.WriteString(fmt.Sprintf(`"%s": %s`, k, jsonToString(v)))
		}
		sb.WriteRune('}')
		return sb.String()
	case []any:
		var sb strings.Builder
		isFirst := true
		for _, v := range t {
			if !isFirst {
				sb.WriteRune(',')
			}
			isFirst = false
			sb.WriteString(jsonToString(v))
		}
		return sb.String()
	default:
		return "null" // might be an error
	}
}
