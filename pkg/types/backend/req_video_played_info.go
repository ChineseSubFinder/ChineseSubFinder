package backend

type ReqVideoPlayedInfo struct {
	PhysicalVideoFileFullPath string `json:"physical_video_file_full_path"` // 视频的物理路径
	SubName                   string `json:"sub_name"`                      // 字幕的名称，不要传递全路径
}
