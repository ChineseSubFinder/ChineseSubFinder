package backend

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"
)

type ReplyJobThings struct {
	JobID     string               `json:"job_id"`
	JobStatus task_queue.JobStatus `json:"job_status"`
	Message   string               `json:"message"`
}
