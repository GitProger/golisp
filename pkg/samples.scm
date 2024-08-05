(define (map-1 f a) 
	(if (null? a) 
		'()
		(cons (f (car a)) (map-1 f (cdr a)))))

(define (cadr l) (car (cdr l)))

(defmacro let (bindings . exprs)
	`((lambda ,(map-1 car bindings) ,@exprs)
		,@(map-1 cadr bindings)))

(let () 1)

(let ((a 1) (b 2)) (+ a b))

(let ((a 1)) (let ((b 2)) (+ a b)))


(define (fn a . b) ((lambda x x) a b))
(fn 1 2 3 4)

((lambda (a . b) ((lambda x x) a b)) 1 2 3 4)


`(+ ,@'(1 2 3))

`((lambda ,(map-1 car '()) 1) ,@(map-1 cadr '()))

`(,'())
`(,@'())

`(10 ,@'() 100)


((lambda (a) ((lambda (b) (+ a b)) 2)) 1)


`((lambda ,(map-1 car '(a)) 1) ,@(map-1 cadr '(1)))
