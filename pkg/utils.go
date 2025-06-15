package lisp

import (
	"reflect"
)

func IsNil(v any) bool {
	return v == nil || reflect.ValueOf(v).IsNil() // reflect.ValueOf(nil).IsNil() -> error
}

func IsNilValue(v any) bool {
	if v == nil {
		return true
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func TypeOf(v any) string { return reflect.TypeOf(v).String() }

func PairOf(v any) Pair {
	if v == nil {
		return nil
	}
	return v.(Pair)
}

func NullSafeCast[T any](v any) T {
	if v == nil {
		var z T
		return z
	} else {
		return v.(T)
	}
}

func ExprOfAny(a any) Expr {
	if c, ok := a.(*ConsCell); ok { // ****
		return Expr{isSExpr: true, sexp: c}
	} else if e, ok := a.(Expr); ok {
		return e
	} else {
		return Expr{atom: a}
	}
}

func AnyFromExpr(e Expr) any {
	if e.isSExpr {
		return e.sexp
	} else {
		return e.atom
	}
}

func MapCons(fn func(any) any, a any) any {
	// if a == nil { return nil }
	if IsNil(a) {
		return a
	}
	return Cons(fn(a.(Pair).Car()), MapCons(fn, a.(Pair).Cdr()))
}

func MapConsUnfold(fn func(any) (any, bool), a any) any {
	// fmt.Printf("Unfold: %+v\n", a)
	if IsNil(a) {
		return a // '() -> '()
	}
	head, unfold := fn(a.(Pair).Car())
	// fmt.Println("res:", head, "|", unfold)
	rest := MapConsUnfold(fn, a.(Pair).Cdr())
	if unfold {
		// v.cdr.cdr. ... = rest
		if _, isList := head.(Pair); isList && IsNil(head) {
			return rest
		}

		if it, isList := head.(Pair); isList {
			for !IsNil(it.Cdr()) {
				it = PairOf(it.Cdr())
			}
			it.(*ConsCell).SetCdr(rest)
			return head
		} else {
			if !IsEmptyList(rest) {
				panic(ExecError{"error unpacking smth like (x y . z ...)"})
			}
			return head
		}
	} else {
		return Cons(head, rest)
	}
}

func FoldlCons(fn func(cur, acc any) any, z any, list Pair) any {
	if IsNil(list) { // list == nil - incorrect
		return z
	}
	return FoldlCons(fn, fn(list.Car(), z), PairOf(list.Cdr()))
}

func IterateCons(list Pair, fn func(v any) bool) {
	for a := list; !IsEmptyList(a); a = PairOf(a.Cdr()) {
		if !fn(a.Car()) {
			return
		}
	}
}

func CheckLength(list Pair, req int) bool { // len(list) <= req
	if req == 0 {
		return IsEmptyList(list)
	}
	return CheckLength(PairOf(list.Cdr()), req-1)
}

func UnfoldCons(p Pair) Pair { // (a b c ... (x y z)) -> (a b c ... x y z)
	if IsEmptyList(p.Cdr()) {
		return PairOf(p.Car())
	}
	return Cons(p.Car(), UnfoldCons(PairOf(p.Cdr())))
}

func Finalize(ctx *LocalScope, res any) any {
	for {
		if newRes := ExprOfAny(res).Exec(ctx); newRes != res {
			res = newRes
		} else {
			return res
		}
	}
}

func CopyCons(v Pair) *ConsCell {
	if IsEmptyList(v) {
		return nil
	}
	return Cons(v.Car(), CopyCons(PairOf(v.Cdr())))
}

func DeepcopyCons(v Pair) *ConsCell {
	if IsEmptyList(v) {
		return nil
	}
	fst := v
	if lst, is := fst.Car().(Pair); is {
		fst = DeepcopyCons(lst)
	}
	return Cons(fst, CopyCons(PairOf(v.Cdr())))
}

func InsertInplace(where, what *ConsCell) *ConsCell { // where.car := expand(what); (1 (2 3 4) 5) -> (1 2 3 4 5)
	res := where
	for what != nil {
		where.car = what.car
		if IsEmptyList(what.cdr) {
			break
		}
		where, what = where.cdr.(*ConsCell), what.cdr.(*ConsCell)
	}
	return res
}

func Insert(where, what *ConsCell) *ConsCell {
	return InsertInplace(DeepcopyCons(where), what)
}
