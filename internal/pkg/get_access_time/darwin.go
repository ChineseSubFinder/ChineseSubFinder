//go:build darwin

package get_access_time

import (
	"os"
	"syscall"
	"time"
)

type OneGetAccessTime struct {
}

func (d OneGetAccessTime) GetOSName() string {
	return "darwin"
}

func (d OneGetAccessTime) GetAccessTime(fileName string) (time.Time, error) {

	// return now time and err if file does not exist
	// TODO: change time.Now()
	fi, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return time.Now(), err
	}

	aTime := fi.Sys().(*syscall.Stat_t).Atimespec
	return time.Unix(int64(aTime.Sec), int64(aTime.Nsec)), nil
}
