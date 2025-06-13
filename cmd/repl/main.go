package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"golisp/lib"
	"golisp/parsing"
	lisp "golisp/pkg"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func InterpretFile(scope *lisp.LocalScope, path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	return Interpret(context.Background(), scope, in, nil, false)
}

func echo(out io.Writer) {
	if out != nil {
		fmt.Fprint(out, "> ")
	}
}

// main REPL function
func Interpret(ctx context.Context, scope *lisp.LocalScope, in io.Reader, out io.Writer, catch bool) error {
	rd := bufio.NewReader(in)

	src := parsing.NewFuncSource(rd.ReadRune)
	defer src.Close()

	echo(out)
	parser := lisp.NewSExpParser(src)
	defer parser.Close()

	defer func() {
		r := recover()
		if e, ok := r.(error); ok && errors.Is(e, lisp.ErrEOF) {
			src.Close()
		} else if r != nil {
			panic(r)
		}
	}()

	cmd := 0
	for src.HasNext() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		func() {
			if catch {
				defer func() {
					if r := recover(); r != nil && out != nil {
						fmt.Fprintf(out, "fatal error: %s\n", r)
						echo(out)
					}
				}()
			}

			expr := parser.ParseSExp()
			res := expr.Exec(scope)
			// cmd++
			if out != nil {
				//fmt.Fprintf(out, "$%d = %s:%v\n", cmd, reflect.TypeOf(res), res)
				if res != nil {
					cmd++
					fmt.Fprintf(out, "$%d = %v\n", cmd, res)
				}
			}
			echo(out)
		}()
	}

	return nil
}

// Use rlwrap for history and arrows
func main() {
	logo()

	err := fs.WalkDir(lib.StandardLibrary, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".scm" {
			return nil
		}
		if in, err := lib.StandardLibrary.Open(path); err != nil {
			return err
		} else {
			fmt.Printf("loading %s...\n", path)
			return Interpret(context.Background(), lisp.Global, in, nil, false)
		}
	})

	if err != nil {
		fmt.Println("error while loading the library: ", err)
		os.Exit(1)
	}

	// _ = InterpretFile(lisp.Global, "hello.scm")

	if len(os.Args) < 2 {
		_ = Interpret(context.Background(), lisp.Global, os.Stdin, os.Stdout, true)
	} else {
		for _, fName := range os.Args[1:] {
			if err := InterpretFile(lisp.Global, fName); err != nil {
				os.Exit(1)
			}
		}
	}
}
