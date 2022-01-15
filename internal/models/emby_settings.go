package models

import "gorm.io/gorm"

type EmbySettings struct {
	gorm.Model
	Enable                 bool              // 是否启用
	AddressUrl             string            // 内网服务器的 url
	APIKey                 string            // API key
	MaxRequestVideoNumber  int               // 最大请求获取视频的数量
	SkipWatched            bool              // 是否跳过已经观看的
	MovieDirectoryMapping  map[string]string // 电影目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
	SeriesDirectoryMapping map[string]string // 连续剧目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
}
