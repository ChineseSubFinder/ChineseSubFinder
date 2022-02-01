//go:build linux

package get_access_time

import (
	"os"
	"syscall"
	"time"
)

type OneGetAccessTime struct {
}

func (d OneGetAccessTime) GetOSName() string {
	return "linux"
}

func (d OneGetAccessTime) GetAccessTime(fileName string) (time.Time, error) {

	// return now time and err if file does not exist
	fi, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return time.Now(), err
	}

	aTime := fi.Sys().(*syscall.Stat_t).Atim
	return time.Unix(aTime.Nanoseconds()/1e9, 0), nil
}
