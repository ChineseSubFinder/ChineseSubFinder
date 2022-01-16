package global_value

// util.go
var (
	DefDebugFolder       = ""
	DefTmpFolder         = ""
	DefSubFixCacheFolder = ""
	WantedExtMap         = make(map[string]string) // 人工确认的需要监控的视频后缀名
	DefExtMap            = make(map[string]string) // 内置支持的视频后缀名列表
	CustomVideoExts      = make([]string, 0)       // 用户额外自定义的视频后缀名列表
)