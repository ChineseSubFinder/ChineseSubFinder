package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) CheckPathHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckPathHandler", err)
	}()

	reqCheckPath := backend.ReqCheckPath{}
	err = c.ShouldBindJSON(&reqCheckPath)
	if err != nil {
		return
	}

	if my_util.IsDir(reqCheckPath.Path) == true {
		c.JSON(http.StatusOK, backend.ReplyCheckPath{
			Valid: true,
		})
	} else {
		c.JSON(http.StatusOK, backend.ReplyCheckPath{
			Valid: false,
		})
	}
}

func (cb ControllerBase) CheckEmbyPathHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckEmbyPathHandler", err)
	}()

	reqCheckPath := backend.ReqCheckEmbyPath{}
	err = c.ShouldBindJSON(&reqCheckPath)
	if err != nil {
		return
	}
	// 需要使用 Emby 做列表转换，从发送过来的 emby_media_path 进行推算，拼接 cfs_media_path 地址，然后读取这个文件夹或者视频是否存在
	// 暂定还是以最近的 Emby 视频列表，再去匹配
	//if my_util.IsDir(reqCheckPath.Path) == true {
	//	c.JSON(http.StatusOK, backend.ReplyCheckEmbyPath{
	//		Valid: true,
	//	})
	//} else {
	//	c.JSON(http.StatusOK, backend.ReplyCheckEmbyPath{
	//		Valid: false,
	//	})
	//}
}
