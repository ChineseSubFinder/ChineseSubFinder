package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/gin-gonic/gin"
)

type ControllerBase struct {
	fileDownloader *file_downloader.FileDownloader
}

func NewControllerBase(fileDownloader *file_downloader.FileDownloader) *ControllerBase {
	return &ControllerBase{
		fileDownloader: fileDownloader,
	}
}

func (cb *ControllerBase) ErrorProcess(c *gin.Context, funcName string, err error) {
	if err != nil {
		cb.fileDownloader.Log.Errorln(funcName, err.Error())
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: err.Error()})
	}
}
