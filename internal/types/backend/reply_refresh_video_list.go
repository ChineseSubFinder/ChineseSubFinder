package backend

type ReplyRefreshVideoList struct {
	Status     string `json:"status"` // "status": "running","stopped"
	ErrMessage string `json:"err_message"`
}
