package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/common"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) CheckSubSupplierHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "LoginHandler", err)
	}()
	nowUserInfo := models.UserInfo{}
	err = c.ShouldBindJSON(&nowUserInfo)
	if err != nil {
		return
	}

	found, dbUserInfo, err := dao.GetUserInfo()
	if err != nil {
		return
	}

	if found == false || dbUserInfo.Username != nowUserInfo.Username || dbUserInfo.Password != nowUserInfo.Password {
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Username or Password Error"})
	} else {
		nowAccessToken := my_util.GenerateAccessToken()
		common.SetAccessToken(nowAccessToken)
		c.JSON(http.StatusOK, backend.ReplyLogin{AccessToken: nowAccessToken})
	}
}
