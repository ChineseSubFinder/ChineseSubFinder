package settings

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/local_http_proxy_server"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
)

type AdvancedSettings struct {
	ProxySettings              *ProxySettings     `json:"proxy_settings"`
	TmdbApiSettings            TmdbApiSettings    `json:"tmdb_api_settings"`
	DebugMode                  bool               `json:"debug_mode"`                     // 是否开启调试模式，这个是写入一个特殊的文件来开启日志的 Debug 输出
	SaveFullSeasonTmpSubtitles bool               `json:"save_full_season_tmp_subtitles"` // 保存整季的缓存字幕
	SubTypePriority            int                `json:"sub_type_priority"`              // 字幕下载的优先级，0 是自动，1 是 srt 优先，2 是 ass/ssa 优先
	SubNameFormatter           int                `json:"sub_name_formatter"`             // 字幕命名格式(默认不填写或者超出范围，则为 emby 格式)，0，emby 支持的的格式（AAA.chinese(简英,subhd).ass or AAA.chinese(简英,xunlei).default.ass），1常规格式（兼容性更好，AAA.zh.ass or AAA.zh.default.ass）
	SaveMultiSub               bool               `json:"save_multi_sub"`                 // 保存多个网站的 Top 1 字幕
	CustomVideoExts            []string           `json:"custom_video_exts""`             // 自定义视频扩展名，是在原有基础上新增。
	FixTimeLine                bool               `json:"fix_time_line"`                  // 开启校正字幕时间轴，默认 false
	Topic                      int                `json:"topic"`                          // 搜索结果的时候，返回 Topic N 以内的
	SuppliersSettings          *SuppliersSettings `json:"suppliers_settings"`             // 每个字幕源的设置
	ScanLogic                  *ScanLogic         `json:"scan_logic"`                     // 扫描的逻辑
	TaskQueue                  *TaskQueue         `json:"task_queue"`                     // 任务队列的设置
	DownloadFileCache          *DownloadFileCache `json:"download_file_cache"`            // 下载文件的缓存
}

func NewAdvancedSettings() *AdvancedSettings {
	return &AdvancedSettings{
		ProxySettings:     NewProxySettings(false, "http", local_http_proxy_server.LocalHttpProxyPort, "127.0.0.1", "10809", "", ""),
		TmdbApiSettings:   *NewTmdbApiSettings(false, "", false),
		CustomVideoExts:   make([]string, 0),
		Topic:             common.DownloadSubsPerSite,
		SuppliersSettings: NewSuppliersSettings(),
		ScanLogic:         NewScanLogic(false, false),
		TaskQueue:         NewTaskQueue(),
		DownloadFileCache: NewDownloadFileCache(),
	}
}
