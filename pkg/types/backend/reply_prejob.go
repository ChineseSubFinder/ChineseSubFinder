package backend

type ReplyPreJob struct {
	IsDone                    bool   `json:"is_done"`                       // 是否完成预处理工作
	StageName                 string `json:"stage_name"`                    // 当前的阶段名称
	HotFixStatus              string `json:"hot_fix_status"`                // 热修复的状态
	ChangeSubNameFormatStatus string `json:"change_sub_name_format_status"` // 修改字幕文件名格式的状态
}
