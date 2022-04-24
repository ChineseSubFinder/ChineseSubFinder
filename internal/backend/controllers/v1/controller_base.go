package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ControllerBase struct {
	log        *logrus.Logger
	cronHelper *cron_helper.CronHelper
}

func NewControllerBase(log *logrus.Logger, cronHelper *cron_helper.CronHelper) *ControllerBase {
	return &ControllerBase{
		log:        log,
		cronHelper: cronHelper,
	}
}

func (cb ControllerBase) GetVersion() string {
	return "v1"
}

func (cb *ControllerBase) ErrorProcess(c *gin.Context, funcName string, err error) {
	if err != nil {
		cb.log.Errorln(funcName, err.Error())
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: err.Error()})
	}
}
