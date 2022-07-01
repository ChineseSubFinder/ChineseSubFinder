//go:build windows

package get_access_time

import (
	"os"
	"syscall"
	"time"
)

type OneGetAccessTime struct {
}

func (d OneGetAccessTime) GetOSName() string {
	return "windows"
}

func (d OneGetAccessTime) GetAccessTime(fileName string) (time.Time, error) {

	// return now time and err if file does not exist
	fi, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return time.Now(), err
	}
	// get last access time for different platform
	// https://studygolang.com/topics/6270
	// https://github.com/golang/go/commit/bd75468a089c8ad38bcb1130c4ed7d2703ef85c1
	// https://github.com/golang/go/issues/31735
	aTime := fi.Sys().(*syscall.Win32FileAttributeData).LastAccessTime
	return time.Unix(aTime.Nanoseconds()/1e9, 0), nil
}
