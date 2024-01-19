package backend

type ReqChangePwd struct {
	OrgPwd string `json:"org_pwd" binding:"required"`
	NewPwd string `json:"new_pwd" binding:"required,min=6,max=36"`
}
