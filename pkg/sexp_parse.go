package lisp

import (
	"encoding/binary"
	"golisp/parsing"
	"strconv"
	"strings"
)

// token stream -> lisp AST
/*
s_expression  ::= atomic_symbol | "(" s_expression "." s_expression ")" | list
list          ::= "(" s_expression { s_expression } ")"
atomic_symbol ::= letter atom_part
atom_part     ::= empty | letter atom_part | number atom_part
letter        ::= "a" | "b" | " ..." | "z"
number        ::= "1" | "2" | " ..." | "9"
empty         ::= " "
*/

func nop() {}

type SExpParser struct {
	*parsing.BaseParser
	processing int
}

func NewSExpParser(cs parsing.CharSource) *SExpParser {
	return &SExpParser{BaseParser: parsing.NewBaseParser(cs)}
}

func ParseSExpString(s string) Expr        { return ParseSExp(parsing.NewStringSource(s)) }
func ParseSExp(cs parsing.CharSource) Expr { return NewSExpParser(cs).ParseSExp() }

// func (parser *SExpParser) ParseSExp() Expr {
// 	if result := parser.parseElement(); parser.Eof() {
// 		if sexp, ok := result.(*ConsCell); ok {
// 			return Expr{isSExpr: true, sexp: sexp}
// 		} else {
// 			return Expr{isSExpr: false, atom: result}
// 		}
// 	}
// 	panic("end of S-expression expected")
// }

func (parser *SExpParser) ParseSExp() Expr {
	result := parser.parseElementTop()
	if sexp, ok := result.(*ConsCell); ok {
		return Expr{isSExpr: true, sexp: sexp}
	} else {
		return Expr{isSExpr: false, atom: result}
	}
}

func (parser *SExpParser) Processing() bool { return parser.processing > 0 }

func (parser *SExpParser) parseElementTop() any {
	parser.skipWhitespaces()
	parser.processing++
	defer func() { parser.processing-- }()
	result := parser.parseValue()
	return result
}

func (parser *SExpParser) parseElement() any {
	parser.skipWhitespaces()
	result := parser.parseValue()
	parser.skipWhitespaces()
	return result
}

func (s *SExpParser) parseValue() any {
	switch {
	case s.Take('('):
		return s.parseList(')')
	case s.Take('['):
		return s.parseList(']')
	case s.Take('"'):
		return s.parseString()
	case s.Take(':'):
		return s.parseKeyword()
	case s.Take('\''):
		return Quote(s.parseElement()) // s.parseSymbol()
	case s.Take('`'):
		return Quasiquote(s.parseElement()) // s.parseSymbol()
	case s.Take(','):
		if s.Take('@') {
			return UnquoteSplicing(s.parseElement())
		} else {
			return Unquote(s.parseElement()) // s.parseSymbol()
		}
	default:
		return s.parseAtom()
	}
}

func (parser *SExpParser) parseIdent() Atomic {
	var sb strings.Builder
	good := func() bool {
		return parser.Between('a', 'z') || parser.Between('A', 'Z') || parser.From("!?+-*/_<>=#") // : (keyword may be #:x or ':x)
	}
	if good() {
		sb.WriteRune(parser.TakeNext())
	}
	for good() || parser.Between('0', '9') {
		sb.WriteRune(parser.TakeNext())
	}
	return Atomic(sb.String())
}

func (parser *SExpParser) parseAtom() any {
	if parser.Between('0', '9') {
		return parser.parseNumber()
	} else {
		return parser.parseIdent()
	}
}

var _ = (*SExpParser).parseSymbol

func (parser *SExpParser) parseSymbol() _Symbol {
	return _Symbol(parser.parseIdent())
}

func (parser *SExpParser) parseKeyword() Keyword {
	return Keyword(parser.parseIdent())
}

func (parser *SExpParser) parseList(end rune) *ConsCell {
	dot := false
	var list []any
	for !parser.Take(end) {
		list = append(list, parser.parseElement())
		if parser.Take('.') {
			list = append(list, parser.parseElement())
			dot = true
			parser.Expect(end)
			break
		}
	}
	var last any = nil
	if dot {
		n := len(list) - 1
		list, last = list[:n], list[n]
	}
	cons := last
	for i := len(list) - 1; i >= 0; i-- {
		cons = Cons(list[i], cons)
	}

	if cons == nil {
		return nil
	}
	return cons.(*ConsCell)
}

// Strings
func (parser *SExpParser) parseString() RawString { // parse "([String]")
	var sb strings.Builder
	for !parser.Take('"') {
		if parser.Eof() {
			panic("string unterminated")
		}

		if parser.Take('\\') {

			if parser.escaped(&sb, '"', '"') ||
				parser.escaped(&sb, '\\', '\\') ||
				parser.escaped(&sb, '/', '/') ||
				parser.escaped(&sb, '\b', 'b') ||
				parser.escaped(&sb, '\f', 'f') ||
				parser.escaped(&sb, '\n', 'n') ||
				parser.escaped(&sb, '\r', 'r') ||
				parser.escaped(&sb, '\t', 't') {
				// next char
			} else if parser.Take('u') {
				value := 0 // pv // 0
				for i := 0; i < 4; i++ {
					value <<= 4
					if parser.Between('0', '9') {
						value = parser.nextHex(value, '0')
					} else if parser.Between('a', 'f') {
						value = parser.nextHex(value, 'a'-10)
					} else if parser.Between('A', 'F') {
						value = parser.nextHex(value, 'A'-10)
					} else {
						panic("expected hex digit")
					}
				}
				binary.Write(&sb, binary.BigEndian, int16(value))
			} else {
				panic("Unknown escape character \\" + string(parser.TakeNext()))
			}

		} else {
			sb.WriteRune(parser.TakeNext())
		}
	}
	return RawString(sb.String())
}

func (parser *SExpParser) nextHex(value, delta int) int {
	value += int(parser.TakeNext()) - delta
	return value
}

func (parser *SExpParser) escaped(sb *strings.Builder, character, expected rune) bool {
	var consumed = parser.Take(expected)
	if consumed {
		sb.WriteRune(character)
	}
	return consumed
}

// Numbers
func (parser *SExpParser) parseNumber() Number {
	var sb strings.Builder
	parser.takeInteger(&sb)

	if parser.Take('.') { // fraction
		sb.WriteRune('.')
		parser.takeDigits(&sb)
	}

	if parser.Take('e') || parser.Take('E') { // exponent
		sb.WriteRune('e')
		if parser.Take('+') { // sign of exponent
			// nothing
		} else if parser.Take('-') {
			sb.WriteRune('-')
		}
		parser.takeDigits(&sb)
	}

	val, err := strconv.ParseFloat(sb.String(), 64)
	if err != nil {
		panic("invalid number: " + err.Error())
	}
	return Number(val)
}

func (parser *SExpParser) takeDigits(sb *strings.Builder) {
	for parser.Between('0', '9') {
		sb.WriteRune(parser.TakeNext())
	}
}

func (parser *SExpParser) takeInteger(sb *strings.Builder) {
	if parser.Take('-') {
		sb.WriteRune('-')
	}
	if parser.Take('0') {
		sb.WriteRune('0')
	} else if parser.Between('1', '9') {
		parser.takeDigits(sb)
	} else {
		panic("invalid number")
	}
}

// Whitespaces
func (parser *SExpParser) skipWhitespaces() { // hangs in waiting
	for parser.Take(' ') || parser.Take('\t') || parser.Take('\n') || parser.Take('\r') {
		src := parser.BaseParser.GetSource()
		if !src.HasNext() {
			// wait
			nop()
		}
	}
}
