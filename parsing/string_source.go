package parsing

import (
	"fmt"
	"unicode/utf8"
)

type StringSource struct {
	data   string
	offset int
	pos    int
}

func NewStringSource(s string) *StringSource {
	return &StringSource{data: s}
}

func (ss *StringSource) HasNext() bool {
	return ss.offset < len(ss.data)
}

func (ss *StringSource) Next() rune {
	r, sz := utf8.DecodeRuneInString(ss.data[ss.offset:])
	ss.pos++
	ss.offset += sz
	return r
}

func (ss *StringSource) Error(msg string) error {
	return fmt.Errorf("parse error: %d: %s", ss.pos, msg)
}

func (ss *StringSource) Close() {}
