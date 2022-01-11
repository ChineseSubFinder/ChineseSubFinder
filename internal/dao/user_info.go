package dao

import (
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
)

func GetUserInfo() (bool, models.UserInfo, error) {
	var userInfos []models.UserInfo
	results := GetDb().Find(&userInfos)
	if results == nil || len(userInfos) == 0 {
		log_helper.GetLogger().Infoln("Need Setup For First Time Use.")
		return false, models.UserInfo{}, nil
	}
	if results.Error != nil {
		return false, models.UserInfo{}, results.Error
	}
	if len(userInfos) > 1 {
		log_helper.GetLogger().Warningln("Found UserInfo len > 2 ")
	}
	// 导出第一个用户的信息
	return true, userInfos[0], nil
}
