package settings

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
)

type EmbySettings struct {
	Enable                bool              `json:"enable"`                   // 是否启用
	AddressUrl            string            `json:"address_url"`              // 内网服务器的 url
	APIKey                string            `json:"api_key"`                  // API key
	MaxRequestVideoNumber int               `json:"max_request_video_number"` // 最大请求获取视频的数量
	SkipWatched           bool              `json:"skip_watched"`             // 是否跳过已经观看的
	MoviePathsMapping     map[string]string `json:"movie_paths_mapping"`      // 电影目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
	SeriesPathsMapping    map[string]string `json:"series_paths_mapping"`     // 连续剧目录的映射，一旦 common setting 的目录修改，需要提示用户确认映射
}

func NewEmbySettings() *EmbySettings {
	return &EmbySettings{
		MaxRequestVideoNumber: 500,
		MoviePathsMapping:     make(map[string]string, 0),
		SeriesPathsMapping:    make(map[string]string, 0),
	}
}

func (e EmbySettings) Check() {
	if e.MaxRequestVideoNumber < common.EmbyApiGetItemsLimitMin ||
		e.MaxRequestVideoNumber > common.EmbyApiGetItemsLimitMax {

		e.MaxRequestVideoNumber = common.EmbyApiGetItemsLimitMin
	}
}
