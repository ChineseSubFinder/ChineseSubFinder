package v1

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) SetupHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SetupHandler", err)
	}()

	setupInfo := backend.ReqSetupInfo{}
	err = c.ShouldBindJSON(&setupInfo)
	if err != nil {
		return
	}
	// 只有当用户不存在的时候才能够执行初始化操作
	found, _, err := dao.GetUserInfo()
	if err != nil {
		return
	}
	if found == true {
		// 存在则反馈无需初始化
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "already setup"})
		return
	} else {
		// 需要创建用户，因为上述判断了没有用户存在，所以就默认直接新建了
		re := dao.GetDb().Create(&setupInfo.UserInfo)
		if re == nil {
			err = errors.New(fmt.Sprintf("dao.GetDb().Create return nil"))
			return
		}
		if re.Error != nil {
			err = re.Error
			return
		}
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
		return
	}
}
