package subtitle_best_api

type AskForDownloadReq struct {
	SubSha256     string `form:"sub_sha256"`       // 文件的 SHA256
	DownloadToken string `form:"download_token"`   // 下载令牌，对应具体服务器会相应回复对应的数据
	ApiKey        string `form:"api_key,optional"` // API Key，非必须，可能是某些用户才有的权限
}

type AskForDownloadReply struct {
	Status            int    `json:"status"`              // 0 失败，1 成功，2 放入了队列，根据返回的时间再下载，3 下载队列满了，需要等待
	Message           string `json:"message"`             // 返回的信息，包括成功和失败的原因
	ScheduledUnixTime int64  `json:"scheduled_unix_time"` // 预约的下载时间
}
