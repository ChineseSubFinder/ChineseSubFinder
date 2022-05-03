package backend

type ReqChangeJobStatus struct {
	Id           string `json:"id"`                           // 任务的唯一 ID
	TaskPriority string `json:"task_priority" default:"high"` // 任务的优先级，high or middle or low priority
}
