package lock

import "sync"

// Lock try lock
type Lock struct {
	isLocked bool
	ll       sync.Mutex
}

// NewLock generate a try lock
func NewLock() *Lock {
	var l Lock
	l.isLocked = false
	return &l
}

// Lock try lock, return lock result
func (l *Lock) Lock() bool {
	l.ll.Lock()
	defer l.ll.Unlock()
	if l.isLocked {
		return false
	}
	l.isLocked = true
	return true
}

// Unlock , Unlock the try lock
func (l *Lock) Unlock() {
	l.ll.Lock()
	defer l.ll.Unlock()
	l.isLocked = false
}
