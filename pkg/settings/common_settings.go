package settings

type CommonSettings struct {
	IntervalOrAssignOrCustom int      `json:"interval_or_assign_or_custom"` // 扫描时间是，使用间隔还是指定时间
	ScanInterval             string   `json:"scan_interval"`                // 一轮字幕扫描的间隔
	Threads                  int      `json:"threads"`                      // 同时扫描的并发数
	RunScanAtStartUp         bool     `json:"run_scan_at_start_up"`         // 完成引导设置后，下次运行程序就开始扫描
	MoviePaths               []string `json:"movie_paths"`                  // 电影的目录
	SeriesPaths              []string `json:"series_paths"`                 // 连续剧的目录
	LocalStaticFilePort      string   `json:"local_static_file_port"`       // 本地静态文件的端口，取消
}

func NewCommonSettings() *CommonSettings {
	return &CommonSettings{
		IntervalOrAssignOrCustom: 0,
		ScanInterval:             "@every 6h", // 间隔 6h 进行字幕的扫描 https://pkg.go.dev/github.com/robfig/cron/v3
		Threads:                  1,
		RunScanAtStartUp:         true,
		MoviePaths:               make([]string, 0),
		SeriesPaths:              make([]string, 0),
		LocalStaticFilePort:      "19037",
	}
}
