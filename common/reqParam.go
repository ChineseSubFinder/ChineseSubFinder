package common

// ReqParam 可选择传入的参数
type ReqParam struct {
	UserExtList []string	// 用户确认的视频后缀名支持列表
	SaveMultiSub bool		// 存储每个网站 Top1 的字幕
	DebugMode bool			// 调试标志位
	Threads   int			// 同时并发的线程数（准确来说在go中不是线程，是 goroutine）
	HttpProxy string		// HttpClient 相关
	UserAgent string		// HttpClient 相关
	Referer   string		// HttpClient 相关
	MediaType string		// HttpClient 相关
	Charset   string		// HttpClient 相关
	Topic	  int			// 搜索结果的时候，返回 Topic N 以内的
}
