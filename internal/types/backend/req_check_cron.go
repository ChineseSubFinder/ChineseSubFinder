package backend

type ReqCheckCron struct {
	ScanInterval string `json:"scan_interval"  binding:"required"`
}
