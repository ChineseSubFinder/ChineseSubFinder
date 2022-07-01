package backend

type ReqCheckPath struct {
	Path string `json:"path"  binding:"required"`
}
