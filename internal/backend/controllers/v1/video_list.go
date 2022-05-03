package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) MovieListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "MovieListHandler", err)
	}()

	bok, allJobs, err := cb.cronHelper.DownloadQueue.GetAllJobs()
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyAllJobs{
			AllJobs: make([]task_queue.OneJob, 0),
		})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyAllJobs{
		AllJobs: allJobs,
	})
}

func (cb ControllerBase) SeriesListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SeriesListHandler", err)
	}()

	bok, allJobs, err := cb.cronHelper.DownloadQueue.GetAllJobs()
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyAllJobs{
			AllJobs: make([]task_queue.OneJob, 0),
		})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyAllJobs{
		AllJobs: allJobs,
	})
}
