package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/tmdb_api"
	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckTmdbApiHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckTmdbApiHandler", err)
	}()

	req := tmdb_api.Req{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}
	if req.ApiKey == "" {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "false"})
		return
	}
	tmdbApi, err := tmdb_api.NewTmdbHelper(cb.fileDownloader.Log,
		req.ApiKey,
		cb.fileDownloader.Settings.AdvancedSettings.ProxySettings)
	if err != nil {
		cb.fileDownloader.Log.Errorln("NewTmdbHelper", err)
		return
	}
	if tmdbApi.Alive() == false {
		cb.fileDownloader.Log.Errorln("tmdbApi.Alive() == false")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "false"})
		return
	} else {
		cb.fileDownloader.Log.Infoln("tmdbApi.Alive() == true")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "true"})
		return
	}
}
