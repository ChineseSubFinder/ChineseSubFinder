package task_queue

import "github.com/dgraph-io/badger/v3"

func IsErrOk(err error) bool {
	if err == badger.ErrKeyNotFound {
		return true
	}

	return false
}
