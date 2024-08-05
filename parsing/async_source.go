package parsing

import (
	"fmt"
)

type AsyncSource struct {
	data   chan rune
	closed bool
	pos    int
}

func NewAsyncSource() *AsyncSource {
	return &AsyncSource{data: make(chan rune, 1)}
}

func (as *AsyncSource) Send(r rune) {
	as.data <- r
}

func (as *AsyncSource) HasNext() bool {
	return !as.closed
}

func (as *AsyncSource) Next() rune {
	r, ok := <-as.data
	if !ok {
		panic("async source ended")
	}
	return r
}

func (as *AsyncSource) Error(msg string) error {
	return fmt.Errorf("parse error: %d: %s", as.pos, msg)
}

func (as *AsyncSource) Close() {
	close(as.data)
	as.closed = true
}
