package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"os"
	"path/filepath"
	"strings"
)

// Restore 从备份还原自动校正的字幕文件
func Restore(movieDir, seriesDir string) error {
	// 搜索出所有的 csf-bk 文件
	backUpSubMoviesFilePathList, err := searchBackUpSubFile(movieDir)
	if err != nil {
		return err
	}
	backUpSubSeriesFilePathList, err := searchBackUpSubFile(seriesDir)
	if err != nil {
		return err
	}
	allBkFilesPath := make([]string, len(backUpSubMoviesFilePathList)+len(backUpSubSeriesFilePathList))
	allBkFilesPath = append(allBkFilesPath, backUpSubMoviesFilePathList...)
	allBkFilesPath = append(allBkFilesPath, backUpSubSeriesFilePathList...)
	// 通过这些文件，判断当前每个 bk 下面是否有相应的文件，如果在则删除，然后再重命名 bk 文件回原来的文件名称
	// Fargo - S04E04 - The Pretend War WEBDL-1080p.chinese(简英,shooter).default.ass.csf-bk
	// Fargo - S04E04 - The Pretend War WEBDL-1080p.chinese(简英,shooter).default.ass
	for index, oneBkFile := range allBkFilesPath {

		fixedFileName := strings.ReplaceAll(oneBkFile, BackUpExt, "")
		if my_util.IsFile(fixedFileName) == true {
			err = os.Remove(fixedFileName)
			if err != nil {
				return err
			}
			err = os.Rename(oneBkFile, fixedFileName)
			if err != nil {
				return err
			}
			log_helper.GetLogger().Infoln("Restore", index, fixedFileName)
		}
	}

	return nil
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
