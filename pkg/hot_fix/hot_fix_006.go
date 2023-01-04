package hot_fix

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"gorm.io/gorm"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/sirupsen/logrus"
)

/*
	因为 TMDB ID 是区分 Movie 和 Series 的，所以···需要把本地的缓存清空
*/
type HotFix006 struct {
	log *logrus.Logger
}

func NewHotFix006(log *logrus.Logger) *HotFix006 {
	return &HotFix006{log: log}
}

func (h HotFix006) GetKey() string {
	return "006"
}

func (h HotFix006) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix006) process() (bool, error) {

	// 查询所有的 IMDB info 出来，把 TMDB ID 设置为 空，需要走重写获取的逻辑
	var imdbInfos []models.IMDBInfo
	dao.GetDb().Find(&imdbInfos)
	err := dao.GetDb().Transaction(func(tx *gorm.DB) error {
		for _, info := range imdbInfos {
			info.TmdbId = ""
			tx.Save(&info)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	// media_infos 表需要全部查询出来删除
	var mediaInfos []models.MediaInfo
	dao.GetDb().Find(&mediaInfos)
	err = dao.GetDb().Transaction(func(tx *gorm.DB) error {
		for _, info := range mediaInfos {
			tx.Delete(&info)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
