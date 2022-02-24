package backend

type ReqRunningLog struct {
	TheLastFewTimes int `json:"the_last_few_times"` // 获取最后几次的运行日志，每次指的是一次字幕的扫描
}
