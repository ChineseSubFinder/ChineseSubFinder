package backend

type ReqChangePwd struct {
	OrgPwd string `json:"org_pwd" binding:"required,min=6,max=12"`
	NewPwd string `json:"new_pwd" binding:"required,min=6,max=12"`
}
