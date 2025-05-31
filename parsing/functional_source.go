package parsing

import (
	"fmt"
	"io"
)

type RuneEmitter func() (rune, int, error)

type FuncRuneSource struct {
	emit RuneEmitter
	done bool
	pos  int
}

func NewFuncSource(fn RuneEmitter) *FuncRuneSource {
	return &FuncRuneSource{emit: fn}
}

func (fs *FuncRuneSource) HasNext() bool {
	return !fs.done
}

func (fs *FuncRuneSource) Next() rune {
	r, _, err := fs.emit()
	if err != nil {
		if err != io.EOF {
			panic(err)
		} else {
			fs.done = true
		}
	}
	fs.pos++
	return r
}

func (fs *FuncRuneSource) Error(msg string) error {
	return fmt.Errorf("parse error: %d: %s", fs.pos, msg)
}

func (fs *FuncRuneSource) Close() {
	fs.done = true
}
