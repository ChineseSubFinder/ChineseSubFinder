package v1

import (
	backend2 "github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RefreshMediaServerSubList 刷新媒体服务器的字幕列表
func (cb *ControllerBase) RefreshMediaServerSubList(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RefreshMediaServerSubList", err)
	}()

	err = cb.videoScanAndRefreshHelper.RefreshMediaServerSubList()
	if err != nil {
		cb.log.Errorln("RefreshMediaServerSubList", err)
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}
