package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func SkipFileInfo(l *logrus.Logger, curFile os.DirEntry, fileFullPath string) bool {

	if curFile.IsDir() == true {
		// 排除缓存文件夹，见 #532
		if strings.HasPrefix(curFile.Name(), ".@__thumb") == true {
			l.Debugln("curFile is dir and match `.@__thumb`, skip")
			return true
		}
	}

	// 跳过不符合的文件，比如 MAC OS 下可能有缓存文件，见 #138
	fi, err := curFile.Info()
	if err != nil {
		l.Errorln("curFile.Info:", curFile.Name(), err)
		return true
	}

	// 封面缓存文件夹中的文件都要跳过 .@__thumb  #581
	// 获取这个文件的父级文件夹的名称，然后判断是否是 .@__thumb 开头的
	parentFolderName := filepath.Base(filepath.Dir(fileFullPath))
	if strings.HasPrefix(parentFolderName, ".@__thumb") == true {
		l.Debugln("curFile is in .@__thumb folder, skip")
		return true
	}

	// 软链接问题 #558
	if fi.Size() < 1000 {

		fileInfo, err := os.Lstat(fileFullPath)
		if err != nil {
			l.Errorln("os.Lstat:", fileFullPath, err)
			return true
		}
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			// 确认是软连接
			l.Debugln("curFile is symlink,", fileFullPath)
			//realPath, err := filepath.EvalSymlinks(fileFullPath)
			//if err == nil {
			//	fmt.Println("Path:", realPath)
			//}
		} else {
			l.Debugln("curFile.Size() < 1000:", curFile.Name())
			return true
		}
	}

	if fi.Size() == 4096 && strings.HasPrefix(curFile.Name(), "._") == true {
		l.Debugln("curFile.Size() == 4096 && Prefix Name == ._*", curFile.Name())
		return true
	}
	// 跳过预告片，见 #315
	if strings.HasSuffix(strings.ReplaceAll(curFile.Name(), filepath.Ext(curFile.Name()), ""), "-trailer") == true {
		l.Debugln("curFile Name has -trailer:", curFile.Name())
		return true
	}

	return false
}
