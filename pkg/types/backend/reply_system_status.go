package backend

type ReplySystemStatus struct {
	IsSetup bool   `json:"is_setup"` // 是否进行给初始化设置（引导设置），设置用户名什么的
	Version string `json:"version"`  // 系统的版本 v0.0.0
}
