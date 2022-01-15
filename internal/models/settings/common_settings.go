package settings

type CommonSettings struct {
	UseHttpProxy     bool     `json:"use_http_proxy"`                    // 是否使用 http 代理
	HttpProxyAddress string   `json:"http_proxy_address"`                // Http 代理地址，内网
	ScanInterval     string   `json:"scan_interval" gorm:"default:'6h'"` // 一轮字幕扫描的间隔
	Threads          int      `json:"threads" gorm:"default:'1'"`        // 同时扫描的并发数
	RunScanAtStartUp bool     `json:"run_scan_at_start_up"`              // 完成引导设置后，下次运行程序就开始扫描
	MoviePaths       []string `json:"movie_paths"`                       // 电影的目录
	SeriesPaths      []string `json:"series_paths"`                      // 连续剧的目录
}
