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
	return timeSpecToTime(aTime), nil
}

func timeSpecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}
