package backend

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"
)

type ReplyAllJobs struct {
	AllJobs []task_queue.OneJob `json:"all_jobs"`
}
