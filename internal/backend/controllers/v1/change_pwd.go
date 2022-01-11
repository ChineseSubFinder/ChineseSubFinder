package v1

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/common"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) ChangePwdHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ChangePwdHandler", err)
	}()

	changePwd := backend.ReqChangePwd{}
	err = c.ShouldBindJSON(&changePwd)
	if err != nil {
		return
	}

	found, dbUserInfo, err := dao.GetUserInfo()
	if err != nil {
		return
	}

	if found == false {
		// 找不到用户
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: "Can't Found UserInfo"})
	} else if dbUserInfo.Password != changePwd.OrgPwd {
		// 原始的密码不对
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Org Password Error"})
	} else {
		// 同意修改密码
		dbUserInfo.Password = changePwd.NewPwd
		re := dao.GetDb().Updates(dbUserInfo)
		if re == nil {
			err = errors.New(fmt.Sprintf("dao.GetDb().Updates return nil"))
			return
		}
		if re.Error != nil {
			err = re.Error
			return
		}
		// 修改密码成功后，会清理 AccessToken，强制要求重写登录
		common.SetAccessToken("")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
	}
}
