package global_value

func Init(customExts []string) {

	WantedExtMap = make(map[string]string) // 人工确认的需要监控的视频后缀名
	DefExtMap = make(map[string]string)    // 内置支持的视频后缀名列表
	CustomVideoExts = customExts           // 用户额外自定义的视频后缀名列表
}

// util.go
var (
	AppVersion           = "" // 程序的版本号
	ConfigRootDirFPath   = ""
	DefDebugFolder       = ""
	DefTmpFolder         = ""
	DefRodTmpFolder      = ""
	DefSubFixCacheFolder = ""
	WantedExtMap         = make(map[string]string) // 人工确认的需要监控的视频后缀名
	DefExtMap            = make(map[string]string) // 内置支持的视频后缀名列表
	CustomVideoExts      = make([]string, 0)       // 用户额外自定义的视频后缀名列表
)
