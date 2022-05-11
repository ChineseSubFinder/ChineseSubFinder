package task_queue

type JobStatus int

const (
	Waiting     JobStatus = iota // 任务正在等待处理
	Committed                    // 任务已经提交，这个可能是提交给服务器，然后等待查询下载 Local 的本地任务不会使用这个标注位
	Failed                       // 任务失败了，在允许的范围内依然会允许重试
	Done                         // 任务完成
	Downloading                  // 任务正在下载
	Ignore                       // 任务被忽略，会存在于任务列表，但是不下载
)

func (c JobStatus) String() string {
	switch c {
	case Waiting:
		return "waiting"
	case Committed:
		return "committed"
	case Failed:
		return "failed"
	case Done:
		return "done"
	case Downloading:
		return "downloading"
	case Ignore:
		return "ignore"
	}
	return "N/A"
}
