package ws

type SubDownloadJobInfo struct {
	Status            string `json:"status"`              // "running", "waiting"，不是运行中，就是等待中
	StartedTime       string `json:"started_time"`        // 任务开始的时间
	WorkingUnitIndex  int    `json:"working_unit_index"`  // 正在处理到第几部电影或者连续剧
	UnitCount         int    `json:"unit_count"`          // 一共有多少部电影或者连续剧
	WorkingUnitName   string `json:"working_unit_name"`   // 电影名称，或者连续剧的名称
	WorkingVideoIndex int    `json:"working_video_index"` // 正在处理到第几个视频
	VideoCount        int    `json:"video_count"`         // 一共有几个视频
	WorkingVideoName  string `json:"working_video_name"`  // 电影名称，或者是连续剧中某一季的某一集的名称
}
