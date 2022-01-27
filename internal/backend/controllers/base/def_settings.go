package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) DefSettingsHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "DefSettingsHandler", err)
	}()

	c.JSON(http.StatusOK, settings.NewSettings())
}
