package lisp

import (
	"fmt"
	"golisp/functional"
	"strings"
)

// lambda -> let, do
// if -> cond, case
// lambda -> list

func (e Expr) String() string {
	if e.isSExpr {
		return e.sexp.String()
	} else {
		return toStr(e.atom)
	}
}

func (e Expr) DebugString() string {
	if e.isSExpr {
		return e.sexp.DebugString()
	} else {
		return TypeOf(e.atom) + ":" + toStr(e.atom)
	}
}

type Pair interface {
	Car() any
	Cdr() any
}

type ConsCell struct {
	car, cdr any
}

func IsCons(v any) bool {
	_, ok := v.(*ConsCell)
	return ok
}

var EmptyList *ConsCell = nil

func IsEmptyList(v any) bool {
	return v == nil || v == EmptyList
}

func (c ConsCell) Car() any       { return c.car }
func (c ConsCell) Cdr() any       { return c.cdr }
func Cons(car, cdr any) *ConsCell { return &ConsCell{car: car, cdr: cdr} }

func (c *ConsCell) SetCar(v any) { c.car = v }
func (c *ConsCell) SetCdr(v any) { c.cdr = v }

func ConsList[T any](a ...T) *ConsCell {
	res := make([]ConsCell, len(a))
	for i, v := range a {
		res[i].car = v
		if i != len(a)-1 {
			res[i].cdr = &res[i+1]
		}
	}
	return &res[0]
}

func ConsListDotted[T any](a ...T) any { // *ConsCell | Atomic
	if len(a) == 1 {
		return a[0]
	}

	res := make([]ConsCell, len(a)-1)
	for i := range res {
		res[i].car = a[i]
		if i != len(res)-1 {
			res[i].cdr = &res[i+1]
		} else {
			res[i].cdr = &a[len(a)-1]
		}
	}
	return &res[0]
}

func toStr(v any) string {
	if vs, ok := v.(fmt.Stringer); ok {
		return vs.String()
	} else if v != nil {
		return fmt.Sprintf("%v", v)
	} else {
		return ""
	}
}

func (c *ConsCell) String() string { // in some scheme version also: (. x) is x
	if IsNil(c) { // we are defenitely CONS, so '()', not just generic 'nil'
		return "()"
	}
	var sb strings.Builder
	for {
		sb.WriteString(toStr(c.Car()))
		if p, ok := c.Cdr().(*ConsCell); ok { // nil also can be casted to *ConsCell
			if !IsNil(c.Cdr()) {
				sb.WriteString(" ")
				c = p
			} else {
				break
			}
		} else {
			if c.Cdr() != nil {
				sb.WriteString(" . " + toStr(c.Cdr()))
			}
			break
		}
	}
	return "(" + sb.String() + ")"
}

func (c *ConsCell) DebugString() string { // (. x) -> x
	if IsNil(c) {
		return "NIL:()"
	}
	var sb strings.Builder
	for {
		sb.WriteString(TypeOf(c.Car()) + ":" + toStr(c.Car()))
		if p, ok := c.Cdr().(*ConsCell); ok {
			if !IsNil(c.Cdr()) {
				sb.WriteString(" ")
				c = p
			} else {
				break
			}
		} else {
			if c.Cdr() != nil {
				sb.WriteString(" . " + TypeOf(c.Cdr()) + ":" + toStr(c.Cdr()))
			}
			break
		}
	}
	return "(" + sb.String() + ")"
}

type Array struct{ storage []any }

func (a Array) Car() any { return a.storage[0] }
func (a Array) Cdr() any { return a.storage[1:] }
func (a Array) String() string {
	return "[" + strings.Join(functional.Map(func(v any) string {
		return v.(fmt.Stringer).String()
	}, a.storage), " ") + "]"
}

func ConsToGoList(p Pair) []any {
	var res []any
	for {
		res = append(res, p.Car())
		if IsEmptyList(p.Cdr()) {
			break
		}
		p = p.Cdr().(Pair)
	}
	return res
}

func ConsToGoListSoft(p Pair) []any {
	var res []any
	for {
		res = append(res, p.Car())
		if IsEmptyList(p.Cdr()) {
			break
		}
		var ok bool
		if p, ok = p.Cdr().(Pair); !ok { // @ . x)
			return append(res, p)
		}
	}
	return res
}

func IsList(p Pair) bool {
	ok := true
	for {
		if IsEmptyList(p.Cdr()) {
			return true
		}
		if p, ok = p.Cdr().(Pair); !ok {
			return false
		}
	}
}
