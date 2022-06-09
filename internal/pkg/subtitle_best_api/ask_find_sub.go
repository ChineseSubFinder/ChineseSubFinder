package subtitle_best_api

type AskFindSubReq struct {
	VideoFeature string `form:"video_feature"`    // VideoFeature ID
	ImdbId       string `form:"imdb_id"`          // IMDB ID
	TmdbId       string `form:"tmdb_id"`          // TMDB ID，这里是这个剧集的 TMDB ID 不是这一集的哈
	Season       int    `form:"season"`           // 如果对应的是电影则可能是 0，没有
	Episode      int    `form:"episode"`          // 如果对应的是电影则可能是 0，没有
	FindSubToken string `form:"find_sub_token"`   // 查询令牌，由 Client 生成的 8 位随机字符，不包含特殊字符即可
	ApiKey       string `form:"api_key,optional"` // API Key，非必须，可能是某些用户才有的权限
}

type AskFindSubReply struct {
	Status            int        `json:"status"`              // 0 失败，1 成功，在缓存中有对应的信息，2 放入了队列，根据返回的时间再查询，3 查询队列满了，需要等待
	Message           string     `json:"message"`             // 返回的信息，包括成功和失败的原因
	Subtitle          []Subtitle `json:"subtitle,optional"`   // 如果查询成功，返回的字幕信息
	ScheduledUnixTime int64      `json:"scheduled_unix_time"` // 预约查询的时间
}
