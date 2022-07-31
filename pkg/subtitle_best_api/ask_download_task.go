package subtitle_best_api

type AskDownloadTaskReply struct {
	Status       int          `json:"status"`                 // 0 失败，1 有任务，2 没有任务
	Message      string       `json:"message"`                // 返回的信息，包括成功和失败的原因
	DownloadInfo DownloadInfo `json:"download_info,optional"` // 下载信息
}

type DownloadInfo struct {
	TaskID       string `json:"task_id"`       // 任务id
	VideoFeature string `json:"video_feature"` // VideoFeature ID
	ImdbId       string `json:"imdb_id"`       // IMDB ID
	TmdbId       string `json:"tmdb_id"`       // TMDB ID，这里是这个剧集的 TMDB ID 不是这一集的哈
	Season       int    `json:"season"`        // 如果对应的是电影则可能是 0，没有
	Episode      int    `json:"episode"`       // 如果对应的是电影则可能是 0，没有
	IsMovie      bool   `json:"is_movie"`      // 是否是电影，如果是电影则 season 和 episode 可能是 0，没有
}
