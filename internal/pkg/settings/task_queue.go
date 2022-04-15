package settings

type TaskQueue struct {
	MaxRetryTimes          int `json:"max_retry_times" default:"3"`            // 单个任务失败后，最大重试次数，超过后会降一级
	OneJobTimeOut          int `json:"one_job_time_out" default:"300"`         // 单个任务的超时时间 5 * 60 s
	Interval               int `json:"interval" default:"10"`                  // 任务的间隔，单位 s，这里会有一个限制，不允许太快,然后会做一定的随机时间范围，当前值 x ~ 2*x 之内随机
	ExpirationTime         int `json:"expiration_time"  default:"90"`          // 添加任务后，过期的时间（单位 day），超过后，任务会降级到 Low
	OneSubDownloadInterval int `json:"one_sub_download_interval" default:"12"` // 一个字幕下载的间隔(单位 h)，不然老是一个循环。对比的基准是 OneJob 的 UpdateTime
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{MaxRetryTimes: 3, Interval: 10, ExpirationTime: 90, OneJobTimeOut: 300, OneSubDownloadInterval: 12}
}
