package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) LogoutHandler(c *gin.Context) {

	// 注销
	common.SetAccessToken("")
	c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok, need ReLogin"})
}
