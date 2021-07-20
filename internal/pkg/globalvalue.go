package pkg

// util.go
var (
	defDebugFolder  = ""
	defTmpFolder    = ""
	wantedExtMap    = make(map[string]string) // 人工确认的需要监控的视频后缀名
	defExtMap       = make(map[string]string) // 内置支持的视频后缀名列表
	customVideoExts = make([]string, 0)       // 用户额外自定义的视频后缀名列表
)
