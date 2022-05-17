package filter

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func SkipFileInfo(l *logrus.Logger, curFile os.DirEntry) bool {

	// 跳过不符合的文件，比如 MAC OS 下可能有缓存文件，见 #138
	fi, err := curFile.Info()
	if err != nil {
		l.Errorln("curFile.Info:", curFile.Name(), err)
		return true
	}

	if fi.Size() < 1000 {
		l.Debugln("curFile.Size() < 1000:", curFile.Name())
		return true
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
