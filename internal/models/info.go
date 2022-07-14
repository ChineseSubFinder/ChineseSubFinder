package models

import "gorm.io/gorm"

type Info struct {
	gorm.Model
	Id           string // 当前设备的随机ID
	Version      string // 当前设备的版本
	MediaServer  string // 媒体服务的名称，没有使用则是 None
	EnableShare  bool   // 是否开启了共享功能
	EnableApiKey bool   // 是否开启本地 http api 功能
}
