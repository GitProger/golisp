`(eq? + +)` -> exception
better compare pointers:
```go
global.Set("+", Func{ ...
``` 
to:
```go
global.Set("+", &Func{ ...
``` 
