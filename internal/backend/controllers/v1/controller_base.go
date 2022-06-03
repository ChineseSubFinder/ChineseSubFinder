package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/base"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/lock"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ControllerBase struct {
	log                     *logrus.Logger
	cronHelper              *cron_helper.CronHelper
	StaticFileSystemBackEnd *base.StaticFileSystemBackEnd

	videoScanAndRefreshHelper           *video_scan_and_refresh_helper.VideoScanAndRefreshHelper
	videoScanAndRefreshHelperIsRunning  bool
	videoScanAndRefreshHelperLocker     lock.Lock
	videoScanAndRefreshHelperErrMessage string
}

func NewControllerBase(log *logrus.Logger, cronHelper *cron_helper.CronHelper) *ControllerBase {
	cb := &ControllerBase{
		log:                     log,
		cronHelper:              cronHelper,
		StaticFileSystemBackEnd: base.NewStaticFileSystemBackEnd(log),
		// 这里因为不进行任务的添加，仅仅是扫描，所以 downloadQueue 可以为 nil
		videoScanAndRefreshHelper: video_scan_and_refresh_helper.NewVideoScanAndRefreshHelper(
			sub_formatter.GetSubFormatter(log, cronHelper.Settings.AdvancedSettings.SubNameFormatter),
			cronHelper.FileDownloader, nil),
		videoScanAndRefreshHelperLocker: lock.NewLock(),
	}

	return cb
}

func (cb *ControllerBase) Close() {
	cb.cronHelper.Stop()
	cb.videoScanAndRefreshHelper.Cancel()
	cb.videoScanAndRefreshHelperLocker.Close()
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
