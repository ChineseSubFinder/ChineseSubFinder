//go:build linux

package get_access_time

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
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
	return my_util.Second2Time(aTime.Nanoseconds() / 1e9), nil

}
