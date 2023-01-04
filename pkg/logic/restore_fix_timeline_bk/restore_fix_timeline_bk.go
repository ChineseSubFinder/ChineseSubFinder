package restore_fix_timeline_bk

import (
	"os"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
)

// CheckSpeFile 目标是检测特定的文件，找到后，先删除，返回一个标志位用于后面的逻辑
func CheckSpeFile() (bool, error) {

	nowSpeFileName := getSpeFileName()

	if pkg.IsFile(nowSpeFileName) == false {
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
	return filepath.Join(pkg.GetConfigRootDirFPath(), specialFileName)
}

/*
	识别 config 文件夹下面由这个特殊的文件，就会执行从 csf-bk 文件还原时间轴修复前的字幕文件
	对于 Linux 是 /config 文件夹下
	对于 Windows 是程序根目录下
	对于 MacOS 需要自行实现
*/
const (
	specialFileName = "RestoreFixTimelineBK"
)
