package dao

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

func UpdateInfo(version string, settings *settings.Settings) *models.Info {
	var infos []models.Info
	GetDb().Find(&infos)

	mediaServerName := ""
	if settings.EmbySettings.Enable == true {
		mediaServerName = "Emby"
	} else {
		mediaServerName = "None"
	}
	if len(infos) == 0 {
		// 不存在则新增
		saveInfo := &models.Info{
			Id:           pkg.RandStringBytesMaskImprSrcSB(64),
			MediaServer:  mediaServerName,
			Version:      version,
			EnableShare:  settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled,
			EnableApiKey: settings.ExperimentalFunction.ApiKeySettings.Enabled,
		}
		GetDb().Save(saveInfo)

		return saveInfo
	} else {
		// 存在则更新
		infos[0].Version = version
		infos[0].MediaServer = mediaServerName
		infos[0].EnableShare = settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled
		infos[0].EnableApiKey = settings.ExperimentalFunction.ApiKeySettings.Enabled
		GetDb().Save(&infos[0])
		return &infos[0]
	}
}
