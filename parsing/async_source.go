package parsing

import (
	"container/list"
	"fmt"
	"sync"
)

type AsyncSource struct {
	mu sync.Mutex

	data   chan rune
	closed bool
	pos    int

	queue *list.List
}

func NewAsyncSource() *AsyncSource {
	return &AsyncSource{data: make(chan rune), queue: list.New()}
}

func (as *AsyncSource) Send(r rune) {
	as.data <- r
}

func (as *AsyncSource) HasNext() bool {
	as.mu.Lock()
	if as.queue.Len() > 0 || !as.closed {
		as.mu.Unlock()
		return true
	}
	as.mu.Unlock()

	// as.data already closed
	if r, ok := <-as.data; !ok {
		return false
	} else {
		as.mu.Lock()
		as.queue.PushBack(r)
		as.mu.Unlock()
		return true
	}
}

func (as *AsyncSource) Next() rune {
	as.mu.Lock()
	if as.queue.Len() > 0 {
		res := as.queue.Front()
		as.queue.Remove(res)
		as.pos++
		as.mu.Unlock()
		return res.Value.(rune)
	}
	as.mu.Unlock()

	r, ok := <-as.data
	if !ok {
		panic("async source ended")
	}
	as.pos++
	return r
}

func (as *AsyncSource) Error(msg string) error {
	return fmt.Errorf("parse error: %d: %s", as.pos, msg)
}

func (as *AsyncSource) Close() {
	as.mu.Lock()
	as.closed = true
	as.mu.Unlock()
	as.Send(END) // to prevent when .HasNext() is called after all symbols ended but before .Close() and returns true which will cause panic
	close(as.data)
}
