package models

type ThirdPartSetVideoPlayedInfo struct {
	PhysicalVideoFileFullPath string `gorm:"primaryKey" json:"physical_video_file_full_path"` // 视频的物理路径
	SubName                   string `json:"sub_name"`                                        // 字幕的名称，需要配合视频进行推算其的文件位置
}
