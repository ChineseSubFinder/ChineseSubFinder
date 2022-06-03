package sub_share_center

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// CopySub2Cache 检测原有字幕是否存在，然后放到缓存目录中
func CopySub2Cache(log *logrus.Logger, orgSubFileFPath, imdbID string, year int, lowTrust bool) (bool, string) {

	nowFolderDir, err := my_folder.GetShareFolderByYear(year)
	if err != nil {
		log.Errorln("CheckOrgSubFileExistAndCopy2Cache.GetShareFolderByYear", err)
		return false, ""
	}

	if lowTrust == true {
		// 低可信度的字幕存储位置
		nowFolderDir = filepath.Join(nowFolderDir, "low_trust")
	}

	err = os.MkdirAll(filepath.Join(nowFolderDir, imdbID), os.ModePerm)
	if err != nil {
		log.Errorln("CheckOrgSubFileExistAndCopy2Cache.MkdirAll", err)
		return false, ""
	}

	desSubFileFPath := filepath.Join(nowFolderDir, imdbID, filepath.Base(orgSubFileFPath))
	err = my_util.CopyFile(orgSubFileFPath, desSubFileFPath)
	if err != nil {
		log.Errorln("CheckOrgSubFileExistAndCopy2Cache.CopyFile", err)
		return false, ""
	}

	return true, desSubFileFPath
}

// ClearExpiredFiles 情况过期的字幕文件，比如数据库中没有其的引用，那么就需要清理
func ClearExpiredFiles() {

}
