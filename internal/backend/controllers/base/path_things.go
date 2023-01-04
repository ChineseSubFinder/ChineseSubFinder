package base

import (
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/emby_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckPathHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckPathHandler", err)
	}()

	reqCheckPath := backend2.ReqCheckPath{}
	err = c.ShouldBindJSON(&reqCheckPath)
	if err != nil {
		return
	}

	if pkg.IsDir(reqCheckPath.Path) == true {
		c.JSON(http.StatusOK, backend2.ReplyCheckPath{
			Valid: true,
		})
	} else {
		c.JSON(http.StatusOK, backend2.ReplyCheckPath{
			Valid: false,
		})
	}
}

func (cb *ControllerBase) CheckEmbyPathHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckEmbyPathHandler", err)
	}()

	reqCheckPath := backend2.ReqCheckEmbyPath{}
	err = c.ShouldBindJSON(&reqCheckPath)
	if err != nil {
		return
	}
	// 需要使用 Emby 做列表转换，从发送过来的 emby_media_path 进行推算，拼接 cfs_media_path 地址，然后读取这个文件夹或者视频是否存在
	// 暂定还是以最近的 Emby 视频列表，再去匹配
	emSettings := settings.EmbySettings{
		Enable:                true,
		AddressUrl:            reqCheckPath.AddressUrl,
		APIKey:                reqCheckPath.APIKey,
		MaxRequestVideoNumber: 2000,
		SkipWatched:           false,
		MoviePathsMapping:     make(map[string]string, 0),
		SeriesPathsMapping:    make(map[string]string, 0),
	}

	if reqCheckPath.PathType == "movie" {
		emSettings.MoviePathsMapping[reqCheckPath.CFSMediaPath] = reqCheckPath.EmbyMediaPath
	} else {
		emSettings.SeriesPathsMapping[reqCheckPath.CFSMediaPath] = reqCheckPath.EmbyMediaPath
	}

	emHelper := emby_helper.NewEmbyHelper(cb.fileDownloader.MediaInfoDealers)
	outList, err := emHelper.CheckPath(&emSettings, reqCheckPath.PathType, emSettings.MaxRequestVideoNumber)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyCheckEmbyPath{
		MediaList: outList,
	})
}
