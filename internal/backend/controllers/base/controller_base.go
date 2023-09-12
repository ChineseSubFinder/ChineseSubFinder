package base

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/pre_job"
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/lock"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/gin-gonic/gin"
)

type ControllerBase struct {
	fileDownloader   *file_downloader.FileDownloader
	proxyCheckLocker lock.Lock
	restartSignal    chan interface{}
	preJob           *pre_job.PreJob
}

func NewControllerBase(
	fileDownloader *file_downloader.FileDownloader,
	restartSignal chan interface{},
	preJob *pre_job.PreJob,
) *ControllerBase {
	return &ControllerBase{
		fileDownloader:   fileDownloader,
		proxyCheckLocker: lock.NewLock(),
		restartSignal:    restartSignal,
		preJob:           preJob,
	}
}

func (cb *ControllerBase) ErrorProcess(c *gin.Context, funcName string, err error) {
	if err != nil {
		cb.fileDownloader.Log.Errorln(funcName, err.Error())
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: err.Error()})
	}
}

func (cb *ControllerBase) Close() {
	cb.proxyCheckLocker.Close()
}
