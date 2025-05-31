package lisp

import (
	"sync"
)

func RunAll(fs ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(fs))
	for _, f := range fs {
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
}
