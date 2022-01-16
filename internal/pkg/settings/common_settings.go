package settings

type CommonSettings struct {
	ScanInterval     string   `json:"scan_interval"`        // 一轮字幕扫描的间隔
	Threads          int      `json:"threads"`              // 同时扫描的并发数
	RunScanAtStartUp bool     `json:"run_scan_at_start_up"` // 完成引导设置后，下次运行程序就开始扫描
	MoviePaths       []string `json:"movie_paths"`          // 电影的目录
	SeriesPaths      []string `json:"series_paths"`         // 连续剧的目录
}

func NewCommonSettings() *CommonSettings {
	return &CommonSettings{
		ScanInterval: "6h",
		Threads:      1,
		MoviePaths:   make([]string, 0),
		SeriesPaths:  make([]string, 0),
	}
}
