package main

// https://www.gnu.org/software/emacs/manual/html_node/elisp/Evaluation.html
// https://www.gnu.org/software/emacs/manual/html_node/elisp/Quoting.html
// https://www.gnu.org/software/emacs/manual/html_node/elisp/Self_002dEvaluating-Forms.html

import (
	"bufio"
	"fmt"
	"golisp/parsing"
	lisp "golisp/pkg"
	"os"
	"sync"
)

func logo() {
	fmt.Println("                                                   ")
	fmt.Println("   GGG   OOO   SSS   CCCC H   H EEEEE M   M EEEEE  ")
	fmt.Println("  G     O   O S     C     H   H E     MM MM E      ")
	fmt.Println("  G  GG O   O  SSS  C     HHHHH EEEEE M M M EEEEE  ")
	fmt.Println("  G   G O   O     S C     H   H E     M   M E      ")
	fmt.Println("   GGG   OOO   SSS   CCCC H   H EEEEE M   M EEEEE  ")
	fmt.Println("                                                   ")
}

func errprintf(format string, args ...any) (int, error) {
	return fmt.Fprintf(os.Stderr, format, args...)
}

var _ = cp

func cp(i int) {
	fmt.Println("CHECKPOINT", i)
}

func makeErr(v any) error {
	if err, ok := v.(error); ok {
		return err
	} else {
		return fmt.Errorf("%v", v)
	}
}

// func () any { return f(a, b, c ...) }
func SyncOn(l sync.Locker, proc func() any) (res any, err error) {
	l.Lock()
	defer func() {
		if r := recover(); r != nil {
			l.Unlock()
			err = makeErr(r)
		}
	}()
	res = proc()
	l.Unlock()
	return
}

func main() {
	logo()

	acs := parsing.NewAsyncSource()
	acs.Send(' ') // for parser's BaseParser not to hang on generic_parser.go:16
	parser := lisp.NewSExpParser(acs)

	cmd := 0
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						errprintf("error: %s\n", r)
					}
				}()
				fmt.Print("> ")
				sexp := parser.ParseSExp()
				res := sexp.Exec(lisp.Global)
				cmd++
				fmt.Printf("$%d = %v\n", cmd, res)
			}()
		}
	}()

	scan := bufio.NewScanner(os.Stdin)
	for {
		if !scan.Scan() {
			errprintf("<EOF>")
			break
		}

		if err := scan.Err(); err != nil {
			errprintf("%s", err)
			break
		}

		for _, r := range scan.Text() {
			acs.Send(r)
		}
		acs.Send('\n')
	}
}
