package log_and_notifi

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

var locker sync.Mutex
var nowInfo string

func Infoln(log *logrus.Logger, args ...interface{}) {
	log.Infoln(args...)

	defer locker.Unlock()
	locker.Lock()

	nowInfo = fmt.Sprintln(args...)
}

func GetNowInfo() string {
	defer locker.Unlock()
	locker.Lock()

	return nowInfo
}
