package lisp

import (
	"fmt"
	"golisp/functional"
)

var Global = &LocalScope{defs: make(map[Atomic]any)}

func init() {
	RegisterBasicForms(Global)
}

func Define(ctx *LocalScope, args Pair) { // Pair of Expr
	if args == nil { // (define)
		panic("define: wrong syntax")
	}

	if args.Cdr() == nil { // nil: (define x)
		Define(ctx, Cons(args.Car(), ExprOfAny(nil)))
	} else {
		def, rest := ExprOfAny(args.Car()), args.Cdr()
		if def.isSExpr { // function: (define (fn a b c . d) e ...)
			lambda := Lambda(ctx, ExprOfAny(def.sexp.Cdr()), // def.sexp.c.Cdr() of type '*ConsCell' instead of usual SExpression, see ****
				functional.Map(ExprOfAny, ConsToGoList(PairOf(rest)))...)
			name := def.sexp.Car().(Atomic)
			ctx.Set(name, lambda)
		} else { // value: (define x 1) (define (sum a b) (+ a b))
			value := ExprOfAny(rest.(Pair).Car()).Exec(ctx)
			ctx.Set(def.atom.(Atomic), value)
		}
	}
}

func Lambda(defCtx *LocalScope, argNames Expr, es ...Expr) Func {
	return Func{
		args: argNames,
		code: es,
		fn: func(callCtx *LocalScope, argValues Pair) any {
			newCtx := defCtx.Sub() // use callCtx for dynamic scoping
			cons := argValues

			if argNames.isSExpr { // (lambda (a b . c) ...)
				var args any = argNames.sexp // maybe nil
				for ; IsCons(args) && !IsNil(args); args, cons = args.(Pair).Cdr(), PairOf(cons.Cdr()) {
					newCtx.Set(args.(Pair).Car().(Atomic), cons.Car())
				}

				if !IsCons(args) && args != nil {
					newCtx.Set(args.(Atomic), cons)
					cons = nil
				} else if !IsNil(args) { // cdr not nil (a b c . d) args
					newCtx.Set(args.(Pair).Cdr().(Atomic), cons.Cdr())
					cons = nil
				}

				if cons != nil {
					panic("too many arguments")
				}
			} else if argNames.atom != nil { // (lambda x ...)
				// argNames.atom == nil in case of (lambda () ...) or (lambda nil ...)
				newCtx.Set(argNames.atom.(Atomic), cons)
			}

			var res any
			for _, e := range es {
				res = e.Exec(newCtx)
			}
			return res
		},
	}
}

func RegisterBasicForms(global *LocalScope) {
	global.Set("nil", nil)
	global.Set("true", Boolean(true))
	global.Set("false", Boolean(false))

	global.Set("car", Func{args: ExprOfAny(ConsList[Atomic]("l")),
		fn: func(ls *LocalScope, args Pair) any {
			list := args.Car().(Pair)
			return list.Car()
		},
	})
	global.Set("cdr", Func{args: ExprOfAny(ConsList[Atomic]("l")),
		fn: func(ls *LocalScope, args Pair) any {
			list := args.Car().(Pair)
			return list.Cdr()
		},
	})
	global.Set("cons", Func{args: ExprOfAny(ConsList[Atomic]("a b")),
		fn: func(ls *LocalScope, args Pair) any {
			head, tail := args.Car(), args.Cdr().(Pair).Car()
			return Cons(head, tail)
		},
	})

	global.Set("define", Func{
		macro: true,
		args:  ExprOfAny(ConsList[Atomic]("name", "value")),
		fn: func(ls *LocalScope, args Pair) any {
			Define(ls, args)
			return nil
		},
	})

	global.Set("lambda", Func{macro: true,
		args: ExprOfAny(ConsList[Atomic]("params", "code")),
		fn: func(ls *LocalScope, args Pair) any {
			var es []Expr
			argList, code := args.Car(), args.Cdr().(Pair)
			for ; code != nil; code = PairOf(code.Cdr()) {
				es = append(es, ExprOfAny(code.Car()))
			}
			return Lambda(ls, ExprOfAny(argList), es...)
		},
	})

	global.Set("quote", Func{
		macro: true,
		args:  ExprOfAny(ConsList[Atomic]("list")),
		fn: func(ls *LocalScope, p Pair) any {
			return p.Car() // just do not evaluate it as other macros do
		},
	})

	global.Set("eval", Func{ // better need context
		args: ExprOfAny(ConsList[Atomic]("code")),
		fn: func(ls *LocalScope, p Pair) any {
			return ExprOfAny(p.Car()).Exec(ls)
		},
	})

	global.Set("apply", Func{ // better need context
		args: ExprOfAny(ConsListDotted[Atomic]("fn", "arg", "args")),
		fn: func(ls *LocalScope, p Pair) any {
			// fun := ExprOfAny(p.Car()).Exec(ls).(Func)
			fun := p.Car().(Func)
			return fun.fn(ls, UnfoldCons(p.Cdr().(Pair)))
		},
	})

	global.Set("println", Func{
		args: ExprOfAny(ConsList[Atomic]("str")),
		fn: func(ls *LocalScope, p Pair) any {
			fmt.Println(string(p.Car().(RawString)))
			return nil
		},
	})

	global.Set("debug", Func{
		args: ExprOfAny(ConsList[Atomic]("obj")),
		fn: func(ls *LocalScope, p Pair) any {
			fmt.Println(p.Car().(DebugStringer).DebugString())
			return nil
		},
	})

	global.Set("version", Func{fn: func(ls *LocalScope, p Pair) any { return "1.0" }})

	global.Set("if", Func{
		macro: true,
		args:  ExprOfAny(ConsListDotted[Atomic]("cond", "t", "f")),
		fn: func(ls *LocalScope, p Pair) any { // `(if ,cond ,t ,f)
			cond := p.Car()
			code := p.Cdr().(Pair)
			t, f := code.Car(), code.Cdr()
			if ExprOfAny(cond).Exec(ls).(Boolable).Bool() {
				return ExprOfAny(t).Exec(ls)
			} else if f != nil {
				if f.(Pair).Cdr() != nil {
					panic("if: wrong syntax")
				}
				return ExprOfAny(f.(Pair).Car()).Exec(ls)
			}
			return nil
		},
	})

	global.Set("+", Func{fn: func(ls *LocalScope, p Pair) any {
		return FoldlCons(func(cur, acc any) any {
			return Number(cur.(Number) + acc.(Number))
		}, Number(0), p)
		// var res Number
		// for a := p; a != nil; a = PairOf(a.Cdr()) {
		// 	res += a.Car().(Number)
		// }
		// return res
	}})
	global.Set("*", Func{fn: func(ls *LocalScope, p Pair) any {
		return FoldlCons(func(cur, acc any) any {
			return Number(cur.(Number) * acc.(Number))
		}, Number(1), p)
	}})

	global.Set("-", Func{fn: func(ls *LocalScope, p Pair) any {
		l := ConsToGoList(p)
		res := l[0].(Number)
		if len(l) <= 1 {
			return -res
		} else {
			for i := 1; i < len(l); i++ {
				res -= l[i].(Number)
			}
		}
		return res
	}})

	global.Set("/", Func{fn: func(ls *LocalScope, p Pair) any {
		l := ConsToGoList(p)
		res := l[0].(Number)
		if len(l) <= 1 {
			return 1.0 / res
		} else {
			div := Number(1)
			for i := 1; i < len(l); i++ {
				div *= l[i].(Number)
			}
			res /= Number(div)
		}
		return res
	}})

	global.Set("null?", Func{
		fn: func(ls *LocalScope, p Pair) any { // p - list of args
			return Boolean(IsNil(p.Car()))
		},
	})

	global.Set("=", Func{fn: func(ls *LocalScope, p Pair) any { // numbers only
		a := ConsToGoList(p)
		for i := range a {
			if i > 0 && a[i-1] != a[i] {
				return Boolean(false)
			}
		}
		return Boolean(true)
	}})
	global.Set("<", Func{fn: func(ls *LocalScope, p Pair) any {
		a := ConsToGoList(p)
		for i := range a {
			if i > 0 && a[i-1].(Number) >= a[i].(Number) {
				return Boolean(false)
			}
		}
		return Boolean(true)
	}})
	global.Set("<=", Func{fn: func(ls *LocalScope, p Pair) any {
		a := ConsToGoList(p)
		for i := range a {
			if i > 0 && a[i-1].(Number) > a[i].(Number) {
				return Boolean(false)
			}
		}
		return Boolean(true)
	}})
	global.Set(">", Func{fn: func(ls *LocalScope, p Pair) any {
		a := ConsToGoList(p)
		for i := range a {
			if i > 0 && a[i-1].(Number) <= a[i].(Number) {
				return Boolean(false)
			}
		}
		return Boolean(true)
	}})
	global.Set(">=", Func{fn: func(ls *LocalScope, p Pair) any {
		a := ConsToGoList(p)
		for i := range a {
			if i > 0 && a[i-1].(Number) < a[i].(Number) {
				return Boolean(false)
			}
		}
		return Boolean(true)
	}})

}
