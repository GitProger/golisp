package parsing_test

import (
	"golisp/parsing"
	lisp "golisp/pkg"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannel(t *testing.T) {
	for _, s := range []string{"hello", "привет", "שלום", "hello world", "", " ", "     "} {
		src := parsing.NewAsyncSource()
		var res strings.Builder
		lisp.RunAll(
			func() {
				for _, c := range s {
					src.Send(c)
				}
				src.Close()
			},
			func() {
				for src.HasNext() {
					r := src.Next()
					if r == parsing.END {
						break
					}
					res.WriteRune(r)
					// res.WriteRune(src.Next())
				}
			})
		assert.Equalf(t, s, res.String(), "result expected to be '%s' but is '%s'", s, res.String())
		//t.Logf("ok! '%s' == '%s'", s, res.String())
	}
}
