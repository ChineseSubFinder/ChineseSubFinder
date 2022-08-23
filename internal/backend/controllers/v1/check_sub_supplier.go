package v1

import (
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckSubSupplierHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckSubSupplierHandler", err)
	}()
	//nowUserInfo := settings.UserInfo{}
	//err = c.ShouldBindJSON(&nowUserInfo)
	//if err != nil {
	//	return
	//}
	//
	//found, dbUserInfo, err := dao.GetUserInfo()
	//if err != nil {
	//	return
	//}
	//
	//if found == false || dbUserInfo.Username != nowUserInfo.Username || dbUserInfo.Password != nowUserInfo.Password {
	//	c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Username or Password Error"})
	//} else {
	//	nowAccessToken := my_util.GenerateAccessToken()
	//	common.SetAccessToken(nowAccessToken)
	//	c.JSON(http.StatusOK, backend.ReplyLogin{AccessToken: nowAccessToken})
	//}
}
