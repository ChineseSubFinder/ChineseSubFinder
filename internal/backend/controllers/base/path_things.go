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
