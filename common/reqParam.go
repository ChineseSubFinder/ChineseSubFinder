package common

// ReqParam 可选择传入的参数
type ReqParam struct {
	UserExtList []string	// 用户确认的视频后缀名支持列表
	DebugMode bool			// 调试标志位
	HttpProxy string		// HttpClient 相关
	UserAgent string		// HttpClient 相关
	Referer   string		// HttpClient 相关
	MediaType string		// HttpClient 相关
	Charset   string		// HttpClient 相关
	Topic	  int			// 搜索结果的时候，返回 Topic N 以内的
}
