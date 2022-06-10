package hot_fix

import (
	"os"
	"path/filepath"

	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/sirupsen/logrus"
)

/*
	嗯···之前对于连续剧的一集的解析 Season 和 Episode 的方式是从文件名得到的，最近看到由反馈到削刮之后，命名是 S01.E01，这样的方式
	那么就可能解析不对，现在需要重新改为从 nfo 或者 xml 文件中得到这个信息，就需要删除之前缓存的数据，然后重新上传，不然之前的数据可能有部分是错误的
*/
type HotFix005 struct {
	log *logrus.Logger
}

func NewHotFix005(log *logrus.Logger) *HotFix005 {
	return &HotFix005{log: log}
}

func (h HotFix005) GetKey() string {
	return "005"
}

func (h HotFix005) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix005) process() (bool, error) {

	shareRootDir, err := my_folder.GetShareSubRootFolder()
	if err != nil {
		h.log.Errorln("GetShareSubRootFolder error:", err.Error())
		return false, err
	}

	// 高可信字幕
	var videoInfos []models.VideoSubInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Find(&videoInfos)
	for _, info := range videoInfos {

		delFileFPath := filepath.Join(shareRootDir, info.StoreRPath)
		if my_util.IsFile(delFileFPath) == true {
			err = os.Remove(delFileFPath)
			if err != nil {
				h.log.Errorln("Remove file:", delFileFPath, " error:", err.Error())
				continue
			}
		}
		dao.GetDb().Delete(&info)
	}
	// 低可信字幕
	var lowTrustVideoInfos []models.LowVideoSubInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Find(&lowTrustVideoInfos)
	for _, info := range lowTrustVideoInfos {

		delFileFPath := filepath.Join(shareRootDir, info.StoreRPath)
		if my_util.IsFile(delFileFPath) == true {
			err = os.Remove(delFileFPath)
			if err != nil {
				h.log.Errorln("Remove file:", delFileFPath, " error:", err.Error())
				continue
			}
		}

		dao.GetDb().Delete(&info)
	}

	return true, nil
}
