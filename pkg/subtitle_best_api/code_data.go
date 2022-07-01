package subtitle_best_api

type CodeReqData struct {
	NowTime string `json:"now_time"`
}

type CodeReplyData struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
