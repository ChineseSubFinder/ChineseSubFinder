package backend

type ReplyPreJob struct {
	IsDone           bool     `json:"is_done"`                      // 是否完成预处理工作
	StageName        string   `json:"stage_name"`                   // 当前的阶段名称
	RenameErrResults []string `json:"rename_err_results,omitempty"` // 重命名结果
	GErrorInfo       string   `json:"g_error_info,omitempty"`       // 全局错误信息
	NowProcessInfo   string   `json:"now_process_info,omitempty"`   // 当前处理的信息
}
