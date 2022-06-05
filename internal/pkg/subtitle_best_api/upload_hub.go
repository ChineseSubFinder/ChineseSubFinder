package subtitle_best_api

type AskForUploadReq struct {
	SubSha256 string `json:"sub_sha256"`
	Trusted   bool   `json:"trusted,optional"` // 是否是信任的字幕
	ImdbId    string `form:"imdb_id,optional"` // IMDB ID
	TmdbId    string `form:"tmdb_id,optional"` // TMDB ID，这里是这个剧集的 TMDB ID 不是这一集的哈
	Season    int    `form:"season,optional"`  // 如果对应的是电影则可能是 0，没有
	Episode   int    `form:"episode,optional"` // 如果对应的是电影则可能是 0，没有
}

type AskForUploadReply struct {
	Status            int    `json:"status"` // 0 失败，1 成功，2 放入了队列，根据返回的时间再上传，3 已经存在，无需上传，本地标记上传了，4 上传队列满了，需要等待
	Message           string `json:"message"`
	ScheduledUnixTime int64  `json:"scheduled_unix_time,omitempty"`
}
