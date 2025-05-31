
(define atom atom?)

(define (zero? x) (= x 0))
;(define (null? x) (eq? x '()))
(define (not x) (if (eq? x #f) #t #f))
(define (pair? p) (not (atom? p)))
; (define (1- x) (- x 1)) (define (1+ x) (+ x 1))
(define (list . x) x)

(define (_len l r)
  (if (null? l) r
    (_len (cdr l) (+ r 1))))

(define (_rev l r)
  (if (null? l) r
    (_rev (cdr l) (cons (car l) r))))

; (length '(1 2 3 4 5))
(define (length l) (_len l 0))
; (reverse '(1 2 3))
(define (reverse l) (_rev l '()))

; (flatten '((1 2) ((3 4 5))))
(define (flatten l)
  (cond ((null? l) '())
        ((pair? l) (append (flatten (car l)) (flatten (cdr l))))
        (else (list l))))


(define (_append2 l1 l2)
  (if (null? l1) l2
    (cons (car l1) (_append2 (cdr l1) l2))))

; (append '(1 2 3) '(4 5))
(define (append . ls)
  (if (null? ls) '()
    (_append2 (car ls) (apply append (cdr ls)))))



; (foldl + 10 '(1 2 3))
(define (foldl f val coll)
  (if (null? coll) val
    (foldl f (f val (car coll)) (cdr coll))))

(define (foldr f val coll)
  (if (null? coll) val
    (f (foldr f (car coll) (cdr coll)) val)))

(define (reduce f coll)
  (foldl f (car coll) (cdr coll)))
(define (reducer f coll)
  (foldr f (car coll) (cdr coll)))

(define (const v) (lambda _ v))

; (define-macro (name . args) . code)
(defmacro define-macro (nargs . code)
  `(defmacro ,(car nargs) ,(cdr nargs) 
    ,@code))

(define-macro (begin . es) `((lambda () ,@es))) ;bug
(define-macro (when c . es) `(if ,c (begin ,@es)))
(define-macro (unless c . es) `(if (not ,c) (begin ,@es)))


(define (nth i l)
  (if (zero? i) 
    (car l)
    (nth (- i 1) (cdr l))))

(define (map-1 f a) 
  (if (null? a) 
    '()
    (cons (f (car a)) (map-1 f (cdr a)))))


(define-macro (nif c f t) `(if ,c ,t ,f))

;(defmacro defun (name args . code)
;  `(define ,name (lambda ,args ,@code)))

(defmacro defun (name args . body)
  (cons 'define (cons (cons name args) body)))

; (let ((a 1) (b 2)) (+ a b))
(defmacro let (bindings . exprs)
  `((lambda ,(map-1 car bindings) ,@exprs)
    ,@(map-1 cadr bindings)))

; (let* ((a 1) (b (+ 1 a))) b)
(define-macro (let* bindings . exprs)
  (if (null? bindings)
    `(begin ,@exprs)
    `(let (,(car bindings))
      (let* ,(cdr bindings) ,@exprs))))

; (or #f 1 2)
(define-macro (or . exprs)
  (if (null? exprs) #f
    `(let ((h ,(car exprs)))
      (if h h (or ,@(cdr exprs))))))

; (and 1 #t 5)
(define-macro (and . exprs)
  (if (null? exprs) #t
    (let ((h (car exprs)))
      (if (= 1 (length exprs)) h
        `(if ,h
          (and ,@(cdr exprs))
          ,h)))))

; (some? zero? '(1 2 3 0))
(define (some? predicate seq)
  (if (null? seq) #f
    (or (predicate (car seq)) (some? predicate (cdr seq)))))

;; (map + '(1 2) '(3 4 5) '(6 7 8 9))
(define (map f . colls)
  (if (or (null? colls) (some? null? colls)) '()
    (cons (apply f (map-1 car colls))
          (apply map f (map-1 cdr colls)))))

; must have defaut clause
; (cond ((= 1 2) 20) ((= 2 2) 30) (else 0))
(define-macro (cond . ps)
  (if (not (null? ps))
    `(if ,(caar ps)
      ,(cadar ps)
      (cond ,@(cdr ps)))))
(define else #t)

; (member 1 '((1) 2 3))
(define (member v l)
  (cond ((null? l) #f)
        ((eq? v (car l)) #t)
        (else (member v (cdr l)))))

; (case 10 ((1 2 3) 10) (else 20))
(define-macro (case val . ps)   
  (if (not (null? ps))
    `(let ((lst (caar (quote ,ps))) (res ,(cadar ps)))
      (cond ((eq? lst 'else) res)
            ((member ,val lst) res)
            (#t (case ,val ,@(cdr ps)))))))

; (bind-lists (a b c) '(1 2 3) (+ a b c))  
(define-macro (bind-lists symbols vals . exprs)
  `(apply (lambda ,symbols ,@exprs) 
    ,vals))

; (let** ((a 1) ((b c) (list a 10))) (list a b c))
(define-macro (let** bindings . exprs)
  (if (null? bindings)
    `(begin ,@exprs)
    (let ((c `(let** ,(cdr bindings) ,@exprs)))
      (if (symbol? (caar bindings))
        `(let (,(car bindings)) ,c)
        `(bind-lists ,(caar bindings) ,(cadar bindings) ,c)))))

; (letrec ((a (+ 10 b)) (b 3)) (+ a b))
; (letrec (( fib (lambda (n) (if (< n 2) 1 (+ (fib (- n 1)) (fib (- n 2))))) )) (fib 10))
; opens to: (let ((fib #f)) (begin (set! fib (lambda (n) (if (< n 2) 1 (+ (fib (- n 1)) (fib (- n 2))))))) (begin (fib 10)))
;(define-macro (letrec bindings . exprs)
;  `(let ,(map-1 (lambda (sym) `(,sym #f)) (map-1 car bindings))
;    (begin 
;      ,@(map-1 
;        (lambda (pair) `(set! ,(car pair) ,(cadr pair)))
;        bindings))
;    (begin ,@exprs)))
; (define-macro (letrec* . stuff) `(letrec ,@stuff))

; (letrec (( fib (lambda (n) (if (< n 2) 1 (+ (fib (- n 1)) (fib (- n 2))))) )) (fib 10))
; opens to: (let ((fib #f)) (begin (set! fib (lambda (n) (if (< n 2) 1 (+ (fib (- n 1)) (fib (- n 2)))))) (fib 10)))
(define-macro (letrec bindings . exprs)
  `(let ,(map-1 (lambda (sym) `(,sym #f)) (map-1 car bindings))
    ,@(map-1 
      (lambda (pair) `(set! ,(car pair) ,(cadr pair)))
      bindings)
    ,@exprs))
