package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/lock"

	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"

	"github.com/allanpk716/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/gin-gonic/gin"
)

type ControllerBase struct {
	fileDownloader   *file_downloader.FileDownloader
	proxyCheckLocker lock.Lock
	restartSignal    chan interface{}
}

func NewControllerBase(fileDownloader *file_downloader.FileDownloader, restartSignal chan interface{}) *ControllerBase {
	return &ControllerBase{
		fileDownloader:   fileDownloader,
		proxyCheckLocker: lock.NewLock(),
		restartSignal:    restartSignal,
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
