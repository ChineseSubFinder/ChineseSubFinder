package subtitle_best_api

type AskForUploadReq struct {
	SubSha256 string `json:"sub_sha256"`
}

type AskForUploadReply struct {
	Status            int    `json:"status"` // 0 失败，1 成功，2 放入了队列，根据返回的时间再上传，3 已经存在，无需上传，本地标记上传了，4 上传队列满了，需要等待
	Message           string `json:"message"`
	ScheduledUnixTime int64  `json:"scheduled_unix_time,omitempty"`
}
