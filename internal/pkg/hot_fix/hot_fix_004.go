package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/sirupsen/logrus"
)

/*
	字幕服务器在判断字幕是否包含中文的时候，没有进行 UTF8 的转换，导致故障丢气大量正确字幕
*/
type HotFix004 struct {
	log *logrus.Logger
}

func NewHotFix004(log *logrus.Logger) *HotFix004 {
	return &HotFix004{log: log}
}

func (h HotFix004) GetKey() string {
	return "004"
}

func (h HotFix004) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix004) process() (bool, error) {

	var videoInfos []models.VideoSubInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Find(&videoInfos)
	for _, info := range videoInfos {
		info.IsSend = false
		dao.GetDb().Save(&info)
	}

	return true, nil
}
