package lisp

import (
	"fmt"
	"strconv"
)

type AtomType int

type (
	NilType   struct{}
	Boolean   bool    // true, false
	Number    float64 // 123
	RawString string  // "..."
	Keyword   string  // :kw
	Atomic    string  // atom

	_Symbol string // sym (unused, same as (quote atom) )
)

var Nil NilType
var EmptyList *ConsCell = nil

func (NilType) String() string { return "#nil" }
func (b Boolean) String() string {
	if bool(b) {
		return "#t"
	}
	return "#f"
}
func (n Number) String() string    { return strconv.FormatFloat(float64(n), 'f', -1, 64) }
func (r RawString) String() string { return "\"" + string(r) + "\"" }
func (s _Symbol) String() string   { return "'" + string(s) }
func (k Keyword) String() string   { return ":" + string(k) }
func (a Atomic) String() string    { return string(a) }

func (NilType) Bool() bool     { return false }
func (b Boolean) Bool() bool   { return bool(b) }
func (n Number) Bool() bool    { return n != 0 }
func (r RawString) Bool() bool { return r != "" }
func (a Atomic) Bool() bool    { return true }
func (c *ConsCell) Bool() bool { return c != nil }

type Executor interface {
	Exec(ctx *LocalScope) any
}

type DebugStringer interface {
	DebugString() string
}

func (NilType) Exec(*LocalScope) any     { return Nil }
func (b Boolean) Exec(*LocalScope) any   { return b }
func (n Number) Exec(*LocalScope) any    { return n }
func (r RawString) Exec(*LocalScope) any { return r }
func (s _Symbol) Exec(*LocalScope) any   { return s }
func (k Keyword) Exec(*LocalScope) any   { return k }

func (a Atomic) Exec(ctx *LocalScope) any {
	if val, ok := ctx.Get(a); ok {
		return val
	} else {
		panic(fmt.Errorf("symbol '%s' not found", a))
	}
}

var symCnt = 0

func GenSym(_ *LocalScope, prefix string) Atomic {
	if prefix == "" {
		prefix = "_"
	}
	res := Atomic(prefix + strconv.Itoa(symCnt))
	symCnt++
	return res
}
