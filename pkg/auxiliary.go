package lisp

import (
	"fmt"
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
