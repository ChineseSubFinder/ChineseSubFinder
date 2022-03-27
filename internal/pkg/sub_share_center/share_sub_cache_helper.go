package sub_share_center

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"path/filepath"
)

// CopySub2Cache 检测原有字幕是否存在，然后放到缓存目录中
func CopySub2Cache(orgSubFileFPath string, year int) (bool, string) {

	nowFolderDir, err := my_util.GetShareFolderByYear(year)
	if err != nil {
		log_helper.GetLogger().Errorln("CheckOrgSubFileExistAndCopy2Cache.GetShareFolderByYear", err)
		return false, ""
	}

	desSubFileFPath := filepath.Join(nowFolderDir, filepath.Base(orgSubFileFPath))
	err = my_util.CopyFile(orgSubFileFPath, desSubFileFPath)
	if err != nil {
		log_helper.GetLogger().Errorln("CheckOrgSubFileExistAndCopy2Cache.CopyFile", err)
		return false, ""
	}

	return true, desSubFileFPath
}

// ClearExpiredFiles 情况过期的字幕文件，比如数据库中没有其的引用，那么就需要清理
func ClearExpiredFiles() {

}
