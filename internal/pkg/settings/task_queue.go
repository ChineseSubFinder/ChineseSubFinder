package settings

type TaskQueue struct {
	MaxRetryTimes  int `json:"max_retry_times" default:"3"`   // 单个任务失败后，最大重试次数，超过后会降级为 Low
	Interval       int `json:"interval" default:"30"`         // 任务的间隔，单位 s，这里会有一个限制，不允许太快,然后会做一定的随机时间范围
	ExpirationTime int `json:"expiration_time"  default:"30"` // 添加任务后，过期的时间（单位 day），超过后，任务会降级到 Low
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{MaxRetryTimes: 3, Interval: 30, ExpirationTime: 30}
}
