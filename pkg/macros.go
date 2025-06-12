package lisp

import (
	"fmt"
)

func init() {
	registerMacros(Global)
}

type (
	Quasiquoted     struct{ boxed any }
	Unquoted        struct{ boxed any }
	UnquotedSpliced struct{ boxed any }
)

func (u Unquoted) String() string        { return "," + toStr(u.boxed) } // or (unquote + ... + )
func (u UnquotedSpliced) String() string { return ",@" + toStr(u.boxed) }
func (q Quasiquoted) String() string     { return "`" + toStr(q.boxed) }

func (q Quasiquoted) Exec(ctx *LocalScope) any {
	return q.Substitute(ctx)
}

func checkQuasi(ctx *LocalScope) {
	if val, ok := ctx.Get("$quasi"); !ok || !val.(bool) {
		panic("unquote out of quasiquote expression")
	}
}

func (a Unquoted) Exec(ctx *LocalScope) any {
	checkQuasi(ctx)
	ctx.Set("$quasi", false)
	defer ctx.Set("$quasi", true)
	res := ExprOfAny(a.boxed).Exec(ctx)
	// for, for instance: `(x y z ... ,a ,b); so ,b is correct, but if we set $quasi to false in ,a it fails
	return res
}

func (a UnquotedSpliced) Exec(ctx *LocalScope) any {
	return Unquoted(a).Exec(ctx)
}

func (s Expr) Substitute(ctx *LocalScope) any {
	if s.isSExpr {
		return MapConsUnfold(func(a any) (any, bool) { // q.boxed:(x y z ... ,a ...)
			// fmt.Printf("map : %20s %s\n", TypeOf(a), a)
			switch t := a.(type) {
			case Unquoted:
				return t.Exec(ctx), false
			case UnquotedSpliced:
				// << map .exec(ctx) t.boxed >>
				return t.Exec(ctx), true
			default:
				return ExprOfAny(t).Substitute(ctx), false
			}
		}, s.sexp)
	} else if u, ok := s.atom.(Unquoted); ok {
		return u.Exec(ctx)
	} else {
		return s.atom
	}
}

func (q Quasiquoted) Substitute(ctx *LocalScope) any {
	ctx.Set("$quasi", true)
	if u, ok := q.boxed.(Unquoted); ok {
		return u.Exec(ctx)
	} else {
		return ExprOfAny(q.boxed).Substitute(ctx)
	}
	// if s, ok := q.boxed.(Expr);            ok { return s.Substitute(ctx)
	// } else if u, ok := q.boxed.(Unquoted); ok { return ExprOfAny(u.Exec(ctx))
	// } else { return ExprOfAny(q.boxed) }
}

func Quasiquote(subj any) Quasiquoted { // `(...) / `a  ?
	return Quasiquoted{boxed: subj} // was:  Quasiquoted{boxed: ExprOfAny(subj)}
}

func Unquote(subj any) Unquoted { // ,(...) / ,a
	// return ExprOfAny(subj.Exec(ctx)) // bad, what to do anologically with unquote-splicing?
	return Unquoted{boxed: subj}
}

func UnquoteSplicing(subj any) UnquotedSpliced { // ,@(...) / ,@a
	return UnquotedSpliced{boxed: subj}
}

func Macroexpand(syntax Expr) any {
	if syntax.isSExpr {
		fmt.Println("EXPAND", syntax)
		panic("not quasiquote")
	} else if q, ok := syntax.atom.(Quasiquoted); ok {
		return q.Substitute(Global)
	} else {
		return AnyFromExpr(syntax)
	}
}

func Macro(defCtx *LocalScope, argNames Expr, es ...Expr) Func {
	return Func{
		macro: true,
		args:  argNames,
		code:  es,
		// ast traversal
		fn: func(callCtx *LocalScope, args Pair) any { // like lambda
			newCtx := defCtx.Sub()
			cons := args
			// fmt.Println(argNames, "<-", args)
			// return Macroexpand(ExprOfAny(args)).Exec(newCtx)

			if argNames.isSExpr { // (lambda (a b . c) ...)
				// fmt.Println("Branch 1")
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
				// fmt.Println("Branch 2")
				// argNames.atom is nil in case of (lambda () ...) or (lambda nil ...)
				newCtx.Set(argNames.atom.(Atomic), cons)
			}

			// fmt.Println("macro exec with:") // newCtx
			var res any
			for _, e := range es {
				// fmt.Println("  to execute:", e)
				res = e.Exec(newCtx)
				// fmt.Println("    res ->", res)
			}
			// fmt.Println("built")

			return res.(Executor).Exec(callCtx) // not newCtx because we should evaluate syntax changes in thw main context immediately unlike in `lambda`
		},
	}
}

func Defmacro(ctx *LocalScope, name Atomic, argNames Expr, es ...Expr) {
	ctx.Set(name, Macro(ctx, argNames, es...))
}

func registerMacros(global *LocalScope) {
	global.Set("defmacro", Func{ // (defmacro name (params...) code...)
		macro: true,
		fn: func(ls *LocalScope, p Pair) any {
			// eval generated code from macroexpand
			/// return Macroexpand(ExprOfAny(p)).Exec(ls)
			args := PairOf(p.Cdr())
			var es []Expr
			argList := args.Car()
			for code := args.Cdr().(Pair); code != nil; code = PairOf(code.Cdr()) {
				es = append(es, ExprOfAny(code.Car()))
			}
			Defmacro(ls, p.Car().(Atomic), ExprOfAny(argList), es...)
			return nil
		},
	})

	global.Set("macroexpand", Func{
		args: ExprOfAny(ConsList[Atomic]("code")),
		fn: func(ls *LocalScope, p Pair) any {
			fmt.Println(p.Car())
			if e, ok := p.Car().(Expr); ok {
				return Macroexpand(e)
			}
			return Macroexpand(ExprOfAny(p.Car().(*ConsCell)))
		},
	})

	global.Set("unquote", Func{
		macro: true,
		args:  ExprOfAny(ConsList[Atomic]("quasiquoted")),
		fn: func(ls *LocalScope, p Pair) any {
			return Unquoted{ExprOfAny(p.Car().(*ConsCell))}
		},
	})

	global.Set("unquote-splicing", Func{
		macro: true,
		fn: func(ls *LocalScope, p Pair) any {
			return UnquotedSpliced{ExprOfAny(p.Car().(*ConsCell))}
		},
	})

	global.Set("quasiquote", Func{
		macro: true,
		fn: func(ls *LocalScope, p Pair) any {
			return Quasiquote(p.Car())
		},
	})
}
