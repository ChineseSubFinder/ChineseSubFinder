package backend

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"
)

type ReqChangeJobStatus struct {
	Id           string               `json:"id"`                           // 任务的唯一 ID
	TaskPriority string               `json:"task_priority" default:"high"` // 任务的优先级，high or middle or low priority
	JobStatus    task_queue.JobStatus `json:"job_status"`                   // 任务的状态 允许设置 Waiting(0) or Ignore(5)
}
