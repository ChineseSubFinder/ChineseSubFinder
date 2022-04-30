package forced_scan_and_down_sub

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"os"
	"path/filepath"
)

// CheckSpeFile 目标是检测特定的文件，找到后，先删除，返回一个标志位用于后面的逻辑
func CheckSpeFile() (bool, error) {

	nowSpeFileName := getSpeFileName()
	if nowSpeFileName == "" {
		return false, errors.New(fmt.Sprintf(`forced_scan_and_down_sub.getSpeFileName() is empty, not support this OS. 
you needd implement getSpeFileName() in internal/logic/forced_scan_and_down_sub/forced_scan_and_down_sub.go`))
	}
	if my_util.IsFile(nowSpeFileName) == false {
		return false, nil
	}
	// 先删除这个文件，然后再标记执行该逻辑
	err := os.Remove(nowSpeFileName)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getSpeFileName() string {

	return filepath.Join(my_folder.GetConfigRootDirFPath(), specialFileNameWindows)
}

/*
	识别 config 文件夹下面由这个特殊的文件，就会执行强制扫描所有视频文件进行字幕的下载（之前有的字幕会被覆盖）
	对于 Linux 是 /config 文件夹下
	对于 Windows 是程序根目录下
	对于 MacOS 是程序根目录下
*/
const (
	specialFileNameWindows = "ForceFullScanAndDownloadSub"
)
