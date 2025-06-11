1. `nil`-values are not Boolable
example:
```lisp
(if nil 1 2)
```
`fatal error: interface conversion: interface is nil, not lisp.Boolable`  

<no-value> == (nil == '())

FIXED by a stub, fixing all relevant `nil` to `lisp.Nil` will be quite complex as there are plenty of places where the code relies of nil-comparison

also here:
    (define (q . args) (display args))
    (q)
    fatal error: interface conversion: interface is nil, not fmt.Stringer
