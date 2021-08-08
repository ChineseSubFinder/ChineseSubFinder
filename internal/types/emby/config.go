package emby

type EmbyConfig struct {
	Url         string //	Emby 的地址，需要带上端口号 http://192.168.1.2:8089
	ApiKey      string //	相应的 API Key
	LimitCount  int    //	最多获取多少更新的内容
	SkipWatched bool   // 	跳过看过的视频，这里会读取所有 Emby 的 User 看过的列表，默认 false
}
