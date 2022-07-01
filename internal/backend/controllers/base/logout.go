package base

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/common"
	"github.com/gin-gonic/gin"
)

func (cb ControllerBase) LogoutHandler(c *gin.Context) {

	// 注销
	common.SetAccessToken("")
	c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok, need ReLogin"})
}
