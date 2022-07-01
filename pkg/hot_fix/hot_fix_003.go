package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/sirupsen/logrus"
)

/*
	upload_played_video_sub 写出了 bug，本地判断字幕的时候，拼接目录错误，导致错误的标记已发送，下载需要改回来，都标记为未标记
	bok, _, err := ch.scanPlayedVideoSubInfo.SubParserHub.DetermineFileTypeFromFile(filepath.Join(shareRootDir, notUploadedVideoSubInfos[0].StoreRPath))
*/
type HotFix003 struct {
	log *logrus.Logger
}

func NewHotFix003(log *logrus.Logger) *HotFix003 {
	return &HotFix003{log: log}
}

func (h HotFix003) GetKey() string {
	return "003"
}

func (h HotFix003) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix003) process() (bool, error) {

	var videoInfos []models.VideoSubInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Find(&videoInfos)
	for _, info := range videoInfos {
		info.IsSend = false
		dao.GetDb().Save(&info)
	}

	return true, nil
}
