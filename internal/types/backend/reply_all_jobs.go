package backend

import "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"

type ReplyAllJobs struct {
	AllJobs []task_queue.OneJob `json:"all_jobs"`
}
