package lisp

import "fmt"

type SyntaxError struct{ msg string }
type ExecError struct{ msg string }
type UnboundError struct{ symbol Atomic }

func (e SyntaxError) Error() string  { return "syntax error: " + e.msg }
func (e ExecError) Error() string    { return e.msg }
func (u UnboundError) Error() string { return fmt.Sprintf("unbound variable: '%s'", u.symbol) }

var (
	TooManyArguments = ExecError{"too many arguments"}
)
