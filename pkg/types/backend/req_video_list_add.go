package backend

type ReqVideoListAdd struct {
	VideoType                 int    `json:"video_type"`                    // 0 是 movie or 1 是 series
	PhysicalVideoFileFullPath string `json:"physical_video_file_full_path"` // 视频的物理路径
	TaskPriorityLevel         int    `json:"task_priority_level"`           // 任务优先级
	MediaServerInsideVideoID  string `json:"media_server_inside_video_id"`  // 媒体服务器内部视频ID
	IsBluray                  bool   `json:"is_bluray"`                     // 这个偏向于给外部 API 调用的时候传递使用。是否是蓝光，目前只支持电影的蓝光，连续剧没有调试过
}

type ReqVideoSkipInfos struct {
	VideoSkipInfos []VideoSkipInfo `json:"video_skip_infos"` // 视频跳过信息
}

type ReplyVideoSkipInfo struct {
	IsSkips []bool `json:"is_skips"` // 是否跳过
}

type VideoSkipInfo struct {
	VideoType                 int    `json:"video_type"`                    // 0 是 movie or 1 是 series
	PhysicalVideoFileFullPath string `json:"physical_video_file_full_path"` // 视频的物理路径
	IsBluray                  bool   `json:"is_bluray"`                     // 是否是蓝光，目前只支持电影的蓝光，连续剧没有调试过
	IsSkip                    bool   `json:"is_skip"`                       // 是否跳过
}
