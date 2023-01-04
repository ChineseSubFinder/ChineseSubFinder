package base

import (
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/emby_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckEmbySettingsHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckEmbySettingsHandler", err)
	}()

	checkEmbyApi := backend2.ReqCheckEmbyAPI{}
	err = c.ShouldBindJSON(&checkEmbyApi)
	if err != nil {
		return
	}

	emSettings := settings.EmbySettings{
		Enable:                true,
		AddressUrl:            checkEmbyApi.AddressUrl,
		APIKey:                checkEmbyApi.APIKey,
		MaxRequestVideoNumber: 2000,
		SkipWatched:           false,
		MoviePathsMapping:     make(map[string]string, 0),
		SeriesPathsMapping:    make(map[string]string, 0),
	}

	emHelper := emby_helper.NewEmbyHelper(cb.fileDownloader.MediaInfoDealers)
	userList, err := emHelper.EmbyApi.GetUserIdList(&emSettings)
	if err != nil {
		return
	}
	if len(userList.Items) <= 0 {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "your emby api key can't get info from emby server"})
	} else {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	}
}
