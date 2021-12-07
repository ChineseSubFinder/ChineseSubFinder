package types

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
)

type Config struct {
	UseProxy                      bool                                       // 是否启用的代理
	HttpProxy                     string                                     // http 代理地址
	EveryTime                     string                                     // 一轮扫描字幕下载的间隔时间
	DebugMode                     bool                                       // 是否启用 Debug 模式，调试功能
	Threads                       int                                        // 同时并发的线程数（准确来说在go中不是线程，是 goroutine）
	SubTypePriority               int                                        // 字幕下载的优先级，0 是自动，1 是 srt 优先，2 是 ass/ssa 优先
	SubNameFormatter              int                                        // 字幕命名格式(默认不填写或者超出范围，则为 emby 格式)，0，emby 支持的的格式（AAA.chinese(简英,subhd).ass or AAA.chinese(简英,xunlei).default.ass），1常规格式（兼容性更好，AAA.zh.ass or AAA.zh.default.ass）
	WhenSubSupplierInvalidWebHook string                                     // 当字幕网站失效的时候，触发的 webhook 地址，默认是 get
	EmbyConfig                    emby.EmbyConfig                            // Emby API 高阶设置参数
	SaveMultiSub                  bool                                       // 保存多个网站的 Top 1 字幕
	SaveOneSeasonSub              bool                                       // 保存整个季度的字幕
	CustomVideoExts               string                                     // 自定义视频扩展名，多个扩展名用英文逗号分隔。是在原有基础上新增。
	RunAtStartup                  bool                                       // 扫描任务是否在启动程序的时候马上执行 见，https://github.com/allanpk716/ChineseSubFinder/issues/50
	SubTimelineFixerConfig        sub_timeline_fiexer.SubTimelineFixerConfig // 时间轴校正配置信息
	FixTimeLine                   bool                                       // 	开启校正字幕时间轴，默认 false

	MovieFolder  string // 电影文件夹
	SeriesFolder string // 连续剧文件夹
	AnimeFolder  string // 日本动画文件夹，很可能不会实现该功能
}
