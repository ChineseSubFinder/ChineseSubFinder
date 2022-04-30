package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"net/http"
)

func (cb ControllerBase) CheckCronHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckCronHandler", err)
	}()

	checkCron := backend.ReqCheckCron{}
	err = c.ShouldBindJSON(&checkCron)
	if err != nil {
		return
	}

	_, err2 := cron.ParseStandard(checkCron.ScanInterval)
	if err2 != nil {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: err2.Error()})
		return
	} else {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
	}
}
