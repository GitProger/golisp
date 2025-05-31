1. `nil`-values are not Boolable
example:
```lisp
(if nil 1 2)
```
`fatal error: interface conversion: interface is nil, not lisp.Boolable`  

<no-value> == (nil == '())
