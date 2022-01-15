package settings

type EmbySettings struct {
	Enable                 bool              `json:"enable"`                                      // 是否启用
	AddressUrl             string            `json:"addressUrl"`                                  // 内网服务器的 url
	APIKey                 string            `json:"APIKey"`                                      // API key
	MaxRequestVideoNumber  int               `json:"maxRequestVideoNumber"  gorm:"default:'500'"` // 最大请求获取视频的数量
	SkipWatched            bool              `json:"skipWatched"`                                 // 是否跳过已经观看的
	MovieDirectoryMapping  map[string]string `json:"movieDirectoryMapping"`                       // 电影目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
	SeriesDirectoryMapping map[string]string `json:"seriesDirectoryMapping"`                      // 连续剧目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
}
