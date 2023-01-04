package backend

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/log_hub"
)

type ReplyRunningLog struct {
	RecentLogs []log_hub.OnceLog `json:"recent_logs"`
}

func NewReplyRunningLog() *ReplyRunningLog {
	return &ReplyRunningLog{
		RecentLogs: make([]log_hub.OnceLog, 0),
	}
}
