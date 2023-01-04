package sub_timeline_fixer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/sirupsen/logrus"
)

// Restore 从备份还原自动校正的字幕文件
func Restore(log *logrus.Logger, movieDirs, seriesDirs []string) (int, error) {

	var BackUpSubMoviesFilePathList = make([]string, 0)
	var BackUpSubSeriesFilePathList = make([]string, 0)

	for _, dir := range movieDirs {
		// 搜索出所有的 csf-bk 文件
		oneBackUpSubMoviesFilePathList, err := searchBackUpSubFile(dir)
		if err != nil {
			return 0, err
		}
		BackUpSubMoviesFilePathList = append(BackUpSubMoviesFilePathList, oneBackUpSubMoviesFilePathList...)
	}
	for _, dir := range seriesDirs {
		// 搜索出所有的 csf-bk 文件
		oneBackUpSubSeriesFilePathList, err := searchBackUpSubFile(dir)
		if err != nil {
			return 0, err
		}
		BackUpSubSeriesFilePathList = append(BackUpSubSeriesFilePathList, oneBackUpSubSeriesFilePathList...)
	}

	allBkFilesPath := make([]string, len(BackUpSubMoviesFilePathList)+len(BackUpSubSeriesFilePathList))
	allBkFilesPath = append(allBkFilesPath, BackUpSubMoviesFilePathList...)
	allBkFilesPath = append(allBkFilesPath, BackUpSubSeriesFilePathList...)
	// 通过这些文件，判断当前每个 bk 下面是否有相应的文件，如果在则删除，然后再重命名 bk 文件回原来的文件名称
	// Fargo - S04E04 - The Pretend War WEBDL-1080p.chinese(简英,shooter).default.ass.csf-bk
	// Fargo - S04E04 - The Pretend War WEBDL-1080p.chinese(简英,shooter).default.ass
	restoreCount := 0
	for index, oneBkFile := range allBkFilesPath {

		fixedFileName := strings.ReplaceAll(oneBkFile, BackUpExt, "")
		if pkg.IsFile(fixedFileName) == true {
			err := os.Remove(fixedFileName)
			if err != nil {
				return 0, err
			}
			err = os.Rename(oneBkFile, fixedFileName)
			if err != nil {
				return 0, err
			}
			restoreCount++
			log.Infoln("Restore", index, fixedFileName)
		}
	}

	return restoreCount, nil
}

func searchBackUpSubFile(dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := searchBackUpSubFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if filepath.Ext(curFile.Name()) == BackUpExt {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}

	return fileFullPathList, nil
}

const TmpExt = ".csf-tmp"
const BackUpExt = ".csf-bk"
