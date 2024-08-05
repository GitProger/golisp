package lisp

import (
	"fmt"
	"golisp/functional"
	"strings"
	"sync"
)

type (
	LocalScope struct {
		parent *LocalScope // constant
		defs   map[Atomic]any
		mu     sync.RWMutex
	}

	Quoted struct {
		boxed any
	}

	Expr struct {
		isSExpr bool
		sexp    *ConsCell
		atom    any
	}

	Func struct {
		macro bool
		args  Expr
		code  []Expr
		fn    func(*LocalScope, Pair) any
	}
)

func Quote(v any) Quoted {
	return Quoted{v}
}

func (q Quoted) Exec(ctx *LocalScope) any {
	return q.boxed
}

func (q Quoted) String() string {
	return "'" + toStr(q.boxed)
}

func (f Func) String() string {
	code := "<native>"
	if f.code != nil {
		code = strings.Join(functional.Map(Expr.String, f.code), " ")
	}
	args := "()"
	e := Expr{}
	if f.args != e {
		args = toStr(f.args)
	}
	if f.macro {
		return fmt.Sprintf("<macro: (macro %s %s)>", args, code)
	} else {
		return fmt.Sprintf("<lambda: (lambda %s %s)>", args, code)
	}
}

func (l *LocalScope) Sub() *LocalScope {
	return &LocalScope{
		parent: l,
		defs:   make(map[Atomic]any),
	}
}

func (l *LocalScope) Get(name Atomic) (any, bool) {
	l.mu.RLock()
	val, ok := l.defs[name]
	l.mu.RUnlock()
	if ok {
		return val, true
	} else if l.parent != nil {
		return l.parent.Get(name)
	} else {
		return nil, false
	}
}

func (l *LocalScope) Del(name Atomic) bool {
	l.mu.RLock()
	_, ok := l.defs[name]
	l.mu.RUnlock()
	if !ok {
		return false
	} else {
		l.mu.Lock()
		delete(l.defs, name)
		l.mu.Unlock()
		return true
	}
}

func (l *LocalScope) Set(name Atomic, value any) (created bool) {
	l.mu.Lock()
	_, created = l.defs[name]
	l.defs[name] = value
	l.mu.Unlock()
	return
}

func (l *LocalScope) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scope %p deriving %p {\n", l, l.parent))
	for k, v := range l.defs {
		sb.WriteString(fmt.Sprintf("'%s': %v\n", k, v))
	}
	sb.WriteByte('}')
	return sb.String()
}

func (expr *ConsCell) Exec(l *LocalScope) any { // (appl ... args)
	if IsNil(expr) {
		panic("empty list is not valid")
	}

	appl, args := expr.Car(), expr.Cdr()
	if fn, callable := ExprOfAny(appl).Exec(l).(Func); !callable {
		panic(fmt.Errorf(`<%s> of type <%s> is not applicable`, appl, TypeOf(appl)))
	} else {
		if fn.macro {
			return fn.fn(l, PairOf(args))
		} else {
			argsEval := MapCons(func(a any) any { return ExprOfAny(a).Exec(l) }, args)
			return fn.fn(l, PairOf(argsEval))
		}
	}
}

// better would be just make Executor everywhere instead of Expr
// but this realization is more dynamic, in some way, consisering types without
// Executor interface implementation
func (e Expr) Exec(ctx *LocalScope) any {
	if e.isSExpr {
		return e.sexp.Exec(ctx)
	} else {
		if sym, ok := e.atom.(Atomic); ok {
			if val, ok := ctx.Get(sym); ok {
				return val
			} else {
				panic(fmt.Errorf("symbol '%s' not found", sym))
			}
		} else if q, ok := e.atom.(Quoted); ok {
			// fmt.Println(reflect.TypeOf(e.atom), e.atom, "->", reflect.TypeOf(q.boxed), q)
			return q.Exec(ctx)
		} else if exec, ok := e.atom.(Executor); ok {
			return exec.Exec(ctx)
		} else {
			return e.atom
		}
	}
}
