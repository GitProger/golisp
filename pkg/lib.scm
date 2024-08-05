
; (define-macro (name . args) . code)
(defmacro define-macro (nargs . code)
  `(defmacro ,(car nargs) ,(cdr nargs) 
    ,@code))

(define-macro (nif c f t) `(if ,c ,t ,f))


(define (map-1 f a) 
  (if (null? a) 
    '()
    (cons (f (car a)) (map-1 f (cdr a)))))


(defmacro let (bindings . exprs)
  `((lambda ,(map-1 car bindings) ,@exprs)
    ,@(map-1 cadr bindings)))

