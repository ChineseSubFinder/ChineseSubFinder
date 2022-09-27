package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg"

	"github.com/allanpk716/ChineseSubFinder/pkg/lock"

	"github.com/allanpk716/ChineseSubFinder/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/sirupsen/logrus"

	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"

	"github.com/allanpk716/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/gin-gonic/gin"
)

type ControllerBase struct {
	fileDownloader   *file_downloader.FileDownloader
	proxyCheckLocker lock.Lock
	restartSignal    chan interface{}
}

func NewControllerBase(loggerBase *logrus.Logger, restartSignal chan interface{}) *ControllerBase {
	return &ControllerBase{
		fileDownloader: file_downloader.NewFileDownloader(
			cache_center.NewCacheCenter("local_task_queue", settings.GetSettings(), loggerBase),
			random_auth_key.AuthKey{
				BaseKey:  pkg.BaseKey(),
				AESKey16: pkg.AESKey16(),
				AESIv16:  pkg.AESIv16(),
			}),
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
