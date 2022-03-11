package ws

type WSType int

const (
	CommonReply           WSType = iota // 通用回复
	Auth                                // Client 发送登录握手消息
	RunningLog                          // Server 回复 Log 的信息
	SubDownloadJobsStatus               // Server 主动发送的字幕下载任务状态信息
)

func (c WSType) String() string {
	switch c {
	case CommonReply:
		return "common_reply"
	case Auth:
		return "auth"
	case RunningLog:
		return "running_log"
	case SubDownloadJobsStatus:
		return "sub_download_jobs_status"
	}
	return "N/A"
}

type AuthMessage int

const (
	AuthOk AuthMessage = iota
	AuthError
)

func (c AuthMessage) String() string {
	switch c {
	case AuthOk:
		return "auth ok"
	case AuthError:
		return "auth error"
	}
	return "N/A"
}

type RunningLogMessage int

const (
	RunningLogRevOk RunningLogMessage = iota
	RunningLogRevError
)

func (c RunningLogMessage) String() string {
	switch c {
	case RunningLogRevOk:
		return "running log recv ok"
	case RunningLogRevError:
		return "running log recv error"
	}
	return "N/A"
}

type CommonMessage int

const (
	OK CommonMessage = iota
)

func (c CommonMessage) String() string {
	switch c {
	case OK:
		return "ok"
	}
	return "N/A"
}

const (
	Preparing  = "preparing"
	ScanMovie  = "scan-movie"
	ScanSeries = "scan-series"
	Waiting    = "waiting"
)

var CloseThisConnect = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
