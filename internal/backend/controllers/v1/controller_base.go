package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ControllerBase struct {
}

func NewControllerBase() *ControllerBase {
	return &ControllerBase{}
}

func (cb ControllerBase) GetVersion() string {
	return "v1"
}

func (cb *ControllerBase) ErrorProcess(c *gin.Context, funcName string, err error) {
	if err != nil {
		log_helper.GetLogger().Errorln(funcName, err.Error())
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: err.Error()})
	}
}
