package lisp

import (
	"fmt"
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

func Test_exec8(t *testing.T) {
	sexp := ParseSExpString(`(println "Hello!")`)
	sexp.Exec(Global)
}

func Test_exec9(t *testing.T) {
	sexp := ParseSExpString(`(println "Hello!")`)
	sexp.Exec(Global)
}

func Test_exec10(t *testing.T) {
	sexp := ParseSExpString(`(if true (println "Hello!"))`)
	sexp.Exec(Global)
}

func Test_exec11(t *testing.T) {
	sexp := ParseSExpString(`(println (if false "Hello" "Bye"))`)
	sexp.Exec(Global)
}

func Test_exec12(t *testing.T) {
	sexp := ParseSExpString(`(if false (println "Hello") (println "Bye"))`)
	sexp.Exec(Global)
}

func Test_exec13(t *testing.T) {
	sexpF := ParseSExpString(`
		(define fac (lambda (n) 
 		  (if (> n 0)
   		    (* n (fac (- n 1)))
   		    1)))`)
	fmt.Println(sexpF.Exec(Global))
	sexp := ParseSExpString(`(fac 5)`)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_lambda_multi(t *testing.T) {
	sexp := ParseSExpString(`((lambda () 1 (println "intermediate") 3))`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_quote1(t *testing.T) {
	sexp := ParseSExpString(`(if (quote x) (println "Hello") (println "Bye"))`)
	fmt.Println(sexp.DebugString())
	sexp.Exec(Global)
}

func Test_exec_quote2(t *testing.T) {
	sexp := ParseSExpString(`(if 'false (println "Hello") (println "Bye"))`)
	fmt.Println(sexp.DebugString())
	sexp.Exec(Global)
}

func Test_exec_eval1(t *testing.T) {
	sexp := ParseSExpString(`(eval '(println "Hello"))`)
	fmt.Println(sexp.DebugString())
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_eval2(t *testing.T) {
	sexp := ParseSExpString(`(eval (quote (println "Hello")))`)
	fmt.Println(sexp)
	sexp.Exec(Global)
}

func Test_exec_apply1(t *testing.T) {
	sexp := ParseSExpString(`(apply + '(1 2))`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}
func Test_exec_apply2(t *testing.T) {
	sexp := ParseSExpString(`(apply + 1 '(2 3))`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_cons1(t *testing.T) {
	sexp := ParseSExpString(`(cons 1 '(2 3))`)
	fmt.Println(sexp.DebugString())
	fmt.Println(sexp.Exec(Global).(DebugStringer).DebugString())
}
func Test_exec_cons2(t *testing.T) {
	sexp := ParseSExpString(`(cons + '(2 3))`)
	fmt.Println(sexp.DebugString())
	fmt.Println(sexp.Exec(Global).(DebugStringer).DebugString())
}
func Test_exec_car1(t *testing.T) {
	sexp := ParseSExpString(`(car '(1 2 3))`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}
func Test_exec_cdr1(t *testing.T) {
	sexp := ParseSExpString(`(cdr '(1 2 3))`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
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
	defer func() {
		_ = recover()
	}()
	sexp := ParseSExpString(`()`)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_closure_1(t *testing.T) {
	ParseSExpString(`(define ten (lambda () 10))`).Exec(Global)
	sexp := ParseSExpString(`(ten)`)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_closure_2(t *testing.T) {
	sexp := ParseSExpString(`(define (f a) (lambda () a))`)
	sexp.Exec(Global)
	sexp = ParseSExpString(`((f 10))`)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_(t *testing.T) {
	sexp := ParseSExpString(`(define (f) 1)`)
	sexp.Exec(Global)
	sexp = ParseSExpString(`(f)`)
	fmt.Println(sexp.Exec(Global))
}

func Test_exec_closure_x(t *testing.T) {
	sexp := ParseSExpString(`'(1 2 3 nil ())`)
	fmt.Println(sexp.Exec(Global))
}

func Test_macro_1(t *testing.T) {
	sexp := ParseSExpString("`,(+ 1 2)") // 3 (self-evaluated form)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_macro_2(t *testing.T) {
	sexp := ParseSExpString("((lambda (a) `(+ ,a)) 10)") // (+ 10)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_macro_3(t *testing.T) {
	sexp := ParseSExpString("((lambda a `(+ ,@a)) 1 2 3)") // (+ 1 2 3)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

// `(+ 1 ,@'(2 3)) -> (+ 1 2 3)
// `(+ 1 ,@(2 3)) -> error

func Test_macro_4(t *testing.T) {
	sexp := ParseSExpString("((lambda (a) `(+ 1 ,@a)) 1)") // (+ 1 . 1)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_macro_5(t *testing.T) {
	defer func() { fmt.Println(recover()) }()
	sexp := ParseSExpString("((lambda (a) `(+ 1 ,@a 3)) 1)") // error
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_macro_6(t *testing.T) {
	sexp := ParseSExpString("((lambda a `(+ 1 ,@a 5)) 2 3 4)") // (+ 1 2 3 4 5)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_Unfold_1(t *testing.T) {
	sexp := ParseSExpString("`(1 ,@'(2 3))") // (1 2 3)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_Unfold_2(t *testing.T) {
	defer func() { fmt.Println(recover()) }()
	sexp := ParseSExpString("`(1 ,@(2 3))") // err, not applicable
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_Unfold_3(t *testing.T) {
	sexp := ParseSExpString("`(1 ,@'())") // (1)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_Unfold_4(t *testing.T) {
	sexp := ParseSExpString("`(1 ,@'() 2)") // (1 2)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_Unfold_5(t *testing.T) {
	sexp := ParseSExpString("`(,@'())") // () or <nil>
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
}

func Test_fn_x(t *testing.T) {
	sexp := ParseSExpString("(cons 1 ((lambda x x) 2 3 4 5))") // (1 2 3 4 5)
	fmt.Println(sexp.Exec(Global))
}

func Test_fn_1(t *testing.T) {
	ParseSExpString( // '() is SExprssion
		`(define (map-1 f a) 
			(if (null? a) 
				'()
				(cons (f (car a)) (map-1 f (cdr a)))))
		`).Exec(Global)
	ParseSExpString(`(define inc (lambda (x) (+ 1 x)))`).Exec(Global)

	sexp := ParseSExpString(`(map-1 inc '(1 2 3))`)
	fmt.Println(sexp)
}

func Test_fn_2(t *testing.T) {
	sexp := ParseSExpString(`((lambda (_ _ . w) w) 1 2 3)`)
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))
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
	fmt.Println(s)
	res := s.Exec(Global)
	fmt.Println(TypeOf(res), res)

	sexp := ParseSExpString("(let ((a 1) (b 2)) (+ a b))")
	fmt.Println(sexp)
	fmt.Println(sexp.Exec(Global))

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
