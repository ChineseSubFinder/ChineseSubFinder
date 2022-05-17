package backend

type ReqVideoListAdd struct {
	VideoType                 int    `json:"video_type"`                    // 0 是 movie or 1 是 series
	PhysicalVideoFileFullPath string `json:"physical_video_file_full_path"` // 视频的物理路径
	TaskPriorityLevel         int    `json:"task_priority_level"`           // 任务优先级
	MediaServerInsideVideoID  string `json:"media_server_inside_video_id"`  // 媒体服务器内部视频ID
	IsBluray                  bool   `json:"is_bluray"`                     // 是否是蓝光，目前只支持电影的蓝光，连续剧没有调试过
}
