package ws

type WSType int

const (
	CommonReply   WSType = iota // 通用回复
	Auth                        // Client 发送登录握手消息
	GetRunningLog               // Client 获取第一次的 Log 信息
	RunningLog                  // Server 回复 Log 的信息
)

func (c WSType) String() string {
	switch c {
	case CommonReply:
		return "common_reply"
	case Auth:
		return "auth"
	case GetRunningLog:
		return "get_running_log"
	case RunningLog:
		return "running_log"
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
