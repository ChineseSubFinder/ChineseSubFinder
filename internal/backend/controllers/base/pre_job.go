package base

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_and_notifi"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PreJobHandler 系统启动后的预处理工作
func (cb *ControllerBase) PreJobHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreJobHandler", err)
	}()

	outInfo := backend.ReplyPreJob{
		IsDone:    cb.preJob.IsDone(),
		StageName: cb.preJob.GetStageName(),
	}
	if cb.preJob.IsDone() == true {
		// 如果完成了，那么就附加对应的消息
		gErr := cb.preJob.GetGError()
		if gErr != nil {
			outInfo.GErrorInfo = gErr.Error()
		} else {
			outInfo.GErrorInfo = ""
		}
		errFiles := cb.preJob.GetRenameResults().ErrFiles
		// 将 errFiles 转为 []string
		outInfo.RenameErrResults = make([]string, 0)
		for k, _ := range errFiles {
			outInfo.RenameErrResults = append(outInfo.RenameErrResults, k)
		}
	} else {
		outInfo.NowProcessInfo = log_and_notifi.GetNowInfo()
	}

	c.JSON(http.StatusOK, outInfo)
}
