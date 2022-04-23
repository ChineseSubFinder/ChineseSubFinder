package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
)

/*
	因为之前有失误把部分临时功能给发布了，所以之前定义 sha1 作为文件的唯一值，现在觉得要升级到 sha256
	那么之前有的需要进行清理一次，然后才能够正确的执行后续新的 sha256 的逻辑
*/
type HotFix002 struct {
}

func NewHotFix002() *HotFix002 {
	return &HotFix002{}
}

func (h HotFix002) GetKey() string {
	return "002"
}

func (h HotFix002) Process() (interface{}, error) {

	defer func() {
		log_helper.GetLogger().Infoln("Hotfix", h.GetKey(), "End")
	}()

	log_helper.GetLogger().Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix002) process() (bool, error) {

	delSubInfo := func(imdbInfo *models.IMDBInfo, cacheInfo *models.VideoSubInfo) bool {
		err := dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Delete(cacheInfo)
		if err != nil {
			log_helper.GetLogger().Warningln("ScanPlayedVideoSubInfo.Scan", ".Delete Association", cacheInfo.SubName, err)
			return false
		}
		// 继续删除这个对象
		dao.GetDb().Delete(cacheInfo)
		log_helper.GetLogger().Infoln("HotFix 002， Sub Association", cacheInfo.SubName)

		return true
	}
	var imdbInfos []models.IMDBInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Preload("VideoSubInfos").Find(&imdbInfos)
	for _, info := range imdbInfos {

		for _, oneSubInfo := range info.VideoSubInfos {
			delSubInfo(&info, &oneSubInfo)
		}
	}

	return true, nil
}
