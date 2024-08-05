package lisp

import (
	"testing"
)

func Test_basicParse_Number_1(t *testing.T) {
	s := ParseSExpString("123")
	if s.atom.(Number) != Number(123) {
		t.Errorf("not eq: %f", s.atom)
	}
}

func Test_basicParse_String(t *testing.T) {
	for _, test := range []struct {
		val string
		res RawString
	}{
		{`"hello"`, "hello"},
		{`"str\n"`, "str\n"},
		{`"\uf09f\u9982"`, "🙂"},
		{`"\thi🙂\n"`, "\thi🙂\n"},
	} {
		actual := ParseSExpString(test.val)
		if actualStr, ok := actual.atom.(RawString); !ok {
			t.Errorf("not a string: %v", actual)
		} else if actualStr != test.res {
			t.Errorf(`expected "%s", got "%s"`, test.res, actualStr)
		}
	}
}

func Test_basicParse_Keyword(t *testing.T) {
	for _, test := range []struct {
		val string
		res Keyword
	}{
		{`:hello`, "hello"},
		{`:f5`, "f5"},
		{`:привет`, ""},
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if test.res != "" {
						t.Errorf("unexpected exception")
					}
				}
			}()
			actual := ParseSExpString(test.val)
			if actualStr, ok := actual.atom.(Keyword); !ok {
				t.Errorf("not a keyword: %v", actual)
			} else if actualStr != test.res {
				t.Errorf(`expected "%s", got "%s"`, test.res, actualStr)
			}
		}()
	}
}

func Test_basicParse_SExpr(t *testing.T) {
	var (
		code   = "(+ a 10)"
		parsed = Cons(Atomic("+"), Cons(Atomic("a"), Cons(Number(10), nil)))
	)
	sexp := ParseSExpString(code)
	res := sexp.String()
	if res != parsed.String() {
		t.Errorf("Not equal, expected:\n%s\ngot:\n%s\n", parsed.String(), res)
	}
}

func Test_basicParse_SExpr_2(t *testing.T) {
	var (
		code   = "(a b . c)"
		parsed = Cons(Atomic("a"), Cons(Atomic("b"), Atomic("c")))
	)
	sexp := ParseSExpString(code)
	res := sexp.String()
	if res != parsed.String() {
		t.Errorf("Not equal, expected:\n%s\ngot:\n%s\n", parsed.String(), res)
	}
}

var factorial_code = `
(define fac (lambda (n) 
  (if (> n 0)
    (* n (fac (- n 1)))
    1)))`

var gt_n_0 = Cons(Atomic(">"), Cons(Atomic("n"), Cons(float64(0), nil)))
var sub_n_1 = Cons(Atomic("-"), Cons(Atomic("n"), Cons(float64(1), nil)))
var factorial = Cons(
	Atomic("define"), Cons(Atomic("fac"), Cons(Cons(Atomic("lambda"),
		Cons(Cons(Atomic("n"), nil),
			Cons(Cons(Atomic("if"), Cons(gt_n_0,
				Cons(Cons(Atomic("*"), Cons(Atomic("n"),
					Cons(Cons(Atomic("fac"), Cons(sub_n_1, nil)), nil))),
					Cons(Number(1), nil)))), nil))), nil)))

func Test_basicParse_Function_factorial(t *testing.T) {
	sexp := ParseSExpString(factorial_code)
	res := sexp.String()
	if res != factorial.String() {
		t.Errorf("Not equal, expected:\n%s\ngot:\n%s\n", factorial.String(), res)
	}
}

// 	sexp := ParseSExpString(`'(1 2 3 nil ())`)
