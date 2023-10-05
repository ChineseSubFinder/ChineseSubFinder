package lock

import (
	"sync"
	"testing"
)

func TestNewLock(t *testing.T) {
	var l = NewLock()
	var wg sync.WaitGroup
	var counter int
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Lock() == false {
				// log error
				println("lock failed")
				return
			}
			counter++
			println("current counter", counter)
			l.Unlock()
		}()
	}
	wg.Wait()
}
