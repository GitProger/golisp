package lisp

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_exec0(t *testing.T) {
	sexp := ParseSExpString(`(version)`)
	assert.Equal(t, VERSION, sexp.Exec(Global))
}

func Test_exec(t *testing.T) {
	sexp := ParseSExpString(`(+ 1 2)`)
	assert.Equal(t, Number(3), sexp.Exec(Global))
}

func Test_exec2(t *testing.T) {
	sexp := ParseSExpString(`(define a 10)`)
	assert.Equal(t, nil, sexp.Exec(Global))
}

func Test_exec3(t *testing.T) {
	sexp := ParseSExpString(`(lambda (x) (+ x 1))`)
	assert.Equal(t, "<lambda: (lambda (x) (+ x 1))>", toStr(sexp.Exec(Global)))
}

func Test_exec4(t *testing.T) {
	sexp := ParseSExpString(`(define inc (lambda (x) (+ x 1)))`)
	assert.Equal(t, nil, sexp.Exec(Global))
	sexp = ParseSExpString(`(inc 10)`)
	assert.Equal(t, Number(11), sexp.Exec(Global))
}

func Test_exec_native(t *testing.T) {
	sexp := ParseSExpString(`+`)
	assert.Equal(t, `<lambda: (lambda () <native>)>`, toStr(sexp.Exec(Global)))
}

func Test_exec5(t *testing.T) {
	sexp := ParseSExpString(`(define (add2 x) (+ x 2))`)
	assert.Equal(t, nil, sexp.Exec(Global))
	sexp = ParseSExpString(`(add2 10)`)
	assert.Equal(t, Number(12), sexp.Exec(Global))
}

func Test_exec6(t *testing.T) {
	sexp := ParseSExpString(`((lambda (x) (+ x 1)) 5)`)
	assert.Equal(t, Number(6), sexp.Exec(Global))
}

func Test_exec7(t *testing.T) {
	sexp := ParseSExpString(`if`)
	assert.Equal(t, true, sexp.Exec(Global).(Func).macro)
}

func captureOutput(main func(), testOutput func(string)) {
	old := os.Stdout
	defer func() { os.Stdout = old }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	testOutput(buf.String())
}

func testCodeOutput(t *testing.T, code, expect string) {
	captureOutput(func() {
		sexp := ParseSExpString(code)
		assert.Nil(t, sexp.Exec(Global))
	}, func(res string) {
		assert.Equal(t, expect, res)
	})
}

func Test_execPrints(t *testing.T) {
	for _, test := range []struct {
		code, expect string
	}{
		{`(println "Hello!")`, "Hello!"},
		{`(println "")`, ""},
		{`(if true (println "Hello!"))`, "Hello!"},
		{`(println (if false "Hello" "Bye"))`, "Bye"},
		{`(if false (println "Hello") (println "Bye"))`, "Bye"},
	} {
		testCodeOutput(t, test.code, test.expect+"\n")
	}
}

func Test_exec_func(t *testing.T) {
	sexpF := ParseSExpString(`
		(define fac (lambda (n) 
 		  (if (> n 0)
   		    (* n (fac (- n 1)))
   		    1)))`)
	assert.Nil(t, sexpF.Exec(Global))
	sexp := ParseSExpString(`(fac 7)`)
	assert.Equal(t, Number(5040), sexp.Exec(Global))
}

func Test_exec_lambda_multi(t *testing.T) {
	captureOutput(func() {
		sexp := ParseSExpString(`((lambda () 1 (println "intermediate") 3))`)
		assert.Equal(t, Number(3), sexp.Exec(Global))
	}, func(s string) {
		assert.Equal(t, "intermediate\n", s)
	})
}

func Test_exec_quote(t *testing.T) {
	h := RawString("Hello")
	assert.Equal(t, h, ParseSExpString(`(if (quote x) "Hello" (1 2))`).Exec(Global))
	assert.Equal(t, h, ParseSExpString(`(if 'false "Hello" "Bye")`).Exec(Global))
}

func Test_exec_eval(t *testing.T) {
	for _, code := range []string{`(eval '(println "Hello"))`, `(eval (quote (println "Hello")))`} {
		captureOutput(func() { ParseSExpString(code).Exec(Global) }, func(s string) {
			assert.Equal(t, "Hello\n", s)
		})
	}
}

func Test_exec_apply_cons_car(t *testing.T) {
	for _, test := range []struct {
		code   string
		expect any
		l      bool
	}{
		{code: `(apply + '(1 2))`, expect: Number(3)},
		{code: `(apply + 1 '(2 3))`, expect: Number(6)},
		{code: `(cons 1 '(2 3))`, expect: ConsList(Number(1), 2, 3), l: true},
		{code: `(cons + '(2 3))`, expect: Cons(Atomic("+").Exec(Global), ConsList(Number(2), 3)), l: true},
		{code: `(car '(1 2 3))`, expect: Number(1)},
		{code: `(cdr '(1 2 3))`, expect: ConsList(Number(2), 3), l: true},
	} {
		sexp := ParseSExpString(test.code)
		expect, res := test.expect, sexp.Exec(Global)
		if test.l {
			expect, res = toStr(expect), toStr(res)
		}
		assert.Equal(t, expect, res)
	}
}

// func Test_exec_closure_dynamic(t *testing.T) {
// 	sexp := ParseSExpString(`(define a 20)`)
// 	fmt.Println(sexp.Exec(Global))
// 	sexp = ParseSExpString(`(((lambda (a) (lambda (_) a)) 10) 0)`)
// 	fmt.Println(sexp.Exec(Global)) // need 20
// }

// about () parsing:
// (if 1 2 (+ ()))
// error
// (defmacro x (a) 1)
// (if 1 2 (x ()))
// ok

func Test_exec_empty(t *testing.T) {
	// defer func() { _ = recover() }()
	sexp := ParseSExpString(`()`)
	assert.Equal(t, EmptyList, sexp.Exec(Global))
	assert.Equal(t, EmptyList, sexp.sexp)
}

func Test_exec_closure_1(t *testing.T) {
	ParseSExpString(`(define ten (lambda () 10))`).Exec(Global)
	assert.Equal(t, Number(10), ParseSExpString(`(ten)`).Exec(Global))
}

func Test_exec_closure_2(t *testing.T) {
	ParseSExpString(`(define (f a) (lambda () a))`).Exec(Global)
	assert.Equal(t, Number(10), ParseSExpString(`((f 10))`).Exec(Global))
}

func Test_exec_list(t *testing.T) {
	lst := ParseSExpString(`'(1 2 3 nil ())`).Exec(Global)
	res := Cons(Number(1), Cons(Number(2), Cons(Number(3), Cons(Atomic("nil"), Cons(EmptyList, EmptyList)))))
	assert.Equal(t, toStr(res), toStr(lst))
}

func Test_macros_and_unfolds(t *testing.T) {
	for _, test := range []struct {
		code   string
		expect any
	}{
		// macro:
		{"`,(+ 1 2)", Number(3)}, // 3 (self-evaluated form)
		{"((lambda (a) `(+ ,a)) 10)", Cons(Atomic("+"), Cons(Number(10), nil))},       // (+ 10)
		{"((lambda a `(+ ,@a)) 1 2 3)", Cons(Atomic("+"), ConsList[Number](1, 2, 3))}, // (+ 1 2 3)
		// `(+ 1 ,@'(2 3)) -> (+ 1 2 3)
		// `(+ 1 ,@(2 3)) -> error
		{"((lambda (a) `(+ 1 ,@a)) 1)", Cons(Atomic("+"), Cons(Number(1), Number(1)))}, // (+ 1 . 1)

		// unfold:
		{"`(1 ,@'(2 3))", ConsList(Number(1), 2, 3)}, // (1 2 3)
		{"`(1 ,@'())", Cons(Number(1), EmptyList)},   // (1)
		{"`(1 ,@'() 2)", ConsList[Number](1, 2)},     // (1 2)
		{"`(,@'())", EmptyList},                      // () or <nil>
	} {
		sexp := ParseSExpString(test.code)
		assert.Equal(t, toStr(test.expect), toStr(sexp.Exec(Global)))
	}
}

func Test_macro_unpack_in_the_middle(t *testing.T) {
	assert.PanicsWithValue(t, ExecError{"error unpacking smth like (x y . z ...)"}, func() {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.IsType(t, ExecError{}, r)
			panic(r)
		}()
		ParseSExpString("((lambda (a) `(+ 1 ,@a 3)) 1)").Exec(Global)
	})
	assert.Equal(t, toStr(Cons(Atomic("+"), ConsList[Number](1, 2, 3, 4, 5))),
		toStr(ParseSExpString("((lambda a `(+ 1 ,@a 5)) 2 3 4)").Exec(Global))) // (+ 1 2 3 4 5)
	// unfold:
	assert.Panics(t, func() { // err, <2> not applicable
		ParseSExpString("`(1 ,@(2 3))").Exec(Global)
	})
}

func Test_fn_x(t *testing.T) {
	assert.Equal(t, toStr(ConsList[Number](1, 2, 3, 4, 5)),
		toStr(ParseSExpString("(cons 1 ((lambda x x) 2 3 4 5))").Exec(Global))) // (1 2 3 4 5)
}

func Test_fn_1(t *testing.T) {
	ParseSExpString( // '() is SExprssion
		`(define (map-1 f a) 
			(if (null? a) 
				'()
				(cons (f (car a)) (map-1 f (cdr a)))))
		`).Exec(Global)
	ParseSExpString(`(define inc (lambda (x) (+ 1 x)))`).Exec(Global)

	assert.Equal(t, toStr(ConsList[Number](2, 3, 4)), toStr(ParseSExpString(`(map-1 inc '(1 2 3))`).Exec(Global)))
}

func Test_fn_2(t *testing.T) {
	assert.Equal(t, "(3)", toStr(ParseSExpString(`((lambda (_ _ . w) w) 1 2 3)`).Exec(Global)))
}

func Test_macro_def_1(t *testing.T) {
	ParseSExpString(
		`(define (map-1 f a) 
			(if (null? a) 
				'()
				(cons (f (car a)) (map-1 f (cdr a)))))
		`).Exec(Global)

	ParseSExpString(`(define (cadr l) (car (cdr l)))`).Exec(Global)

	ParseSExpString(
		`(defmacro let (bindings . exprs)
			` + "`" + `((lambda ,(map-1 car bindings) ,@exprs)
						 ,@(map-1 cadr bindings)))`).Exec(Global)

	// illustration:
	//  `((lambda ,(map-1 car '((a 1) (b 2))) (+ a b)) ,@(map-1 cadr '((a 1) (b 2)))))
	s := ParseSExpString("`((lambda ,(map-1 car '((a 1) (b 2))) (+ a b)) ,@(map-1 cadr '((a 1) (b 2))))")
	res := s.Exec(Global)
	assert.Equal(t, "((lambda (a b) (+ a b)) 1 2)", toStr(res))

	sexp := ParseSExpString("(let ((a 1) (b 2)) (+ a b))")
	assert.Equal(t, Number(3), sexp.Exec(Global))
}

// func Test_gap_1(t *testing.T) {
// 	s := ParseSExpString("")
// 	fmt.Println(s)
// 	fmt.Println(s.Exec(Global))
// }

// (define a (lambda (
//                    x y)
//    (+ x
//       y)))
