package settings

type TaskQueue struct {
	MaxRetryTimes           int    `json:"max_retry_times" default:"3"`            // 单个任务失败后，最大重试次数，超过后会降一级
	OneJobTimeOut           int    `json:"one_job_time_out" default:"300"`         // 单个任务的超时时间 5 * 60 s
	Interval                int    `json:"interval" default:"10"`                  // 任务的间隔，单位 s，这里会有一个限制，不允许太快,然后会做一定的随机时间范围，当前值 x ~ 2*x 之内随机
	ExpirationTime          int    `json:"expiration_time"  default:"90"`          // 单位天。1. 一个视频的 CreatedTime 在这个时间范围内，都会被下载字幕（除非已经观看跳过启用了）。2. 如果下载失败的任务，AddTime 超过了这个时间，那么就标记为 Failed
	DownloadSubDuringXDays  int    `json:"download_sub_during_x_days" default:"7"` // 如果创建了 x 天，且有内置的中文字幕，那么也不进行下载了
	OneSubDownloadInterval  int    `json:"one_sub_download_interval" default:"12"` // 一个字幕下载的间隔(单位 h)，不然老是一个循环。对比的基准是 OneJob 的 UpdateTime
	CheckPublicIPTargetSite string `json:"check_pulic_ip_target_site" default:""`  // 检测本机外网 IP 的目标地址，必须是返回直接的 IP 字符串，不需要解析。; 分割
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		MaxRetryTimes:           3,
		OneJobTimeOut:           300,
		Interval:                10,
		ExpirationTime:          90,
		DownloadSubDuringXDays:  7,
		OneSubDownloadInterval:  12,
		CheckPublicIPTargetSite: "",
	}
}

func (t *TaskQueue) Check() {
	if t.MaxRetryTimes < 1 || t.MaxRetryTimes > 5 {
		t.MaxRetryTimes = 3
	}
	if t.OneJobTimeOut < 300 || t.OneJobTimeOut > 600 {
		t.OneJobTimeOut = 300
	}
	if t.Interval < 10 || t.Interval > 60 {
		t.Interval = 10
	}
	if t.ExpirationTime < 1 || t.ExpirationTime > 180 {
		t.ExpirationTime = 90
	}
	if t.DownloadSubDuringXDays < 1 || t.DownloadSubDuringXDays > 30 {
		t.DownloadSubDuringXDays = 7
	}
	if t.OneSubDownloadInterval < 12 || t.OneSubDownloadInterval > 48 {
		t.OneSubDownloadInterval = 12
	}
}
