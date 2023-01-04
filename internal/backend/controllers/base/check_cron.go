package base

import (
	"net/http"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func (cb *ControllerBase) CheckCronHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckCronHandler", err)
	}()

	checkCron := backend2.ReqCheckCron{}
	err = c.ShouldBindJSON(&checkCron)
	if err != nil {
		return
	}

	_, err2 := cron.ParseStandard(checkCron.ScanInterval)
	if err2 != nil {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: err2.Error()})
		return
	} else {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	}
}
