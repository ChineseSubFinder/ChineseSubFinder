package v1

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	TTaskqueue "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AddJobHandler 外部 API 接口添加任务的处理
func (cb ControllerBase) AddJobHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "AddJobHandler", err)
	}()

	videoListAdd := backend.ReqVideoListAdd{}
	err = c.ShouldBindJSON(&videoListAdd)
	if err != nil {
		return
	}
	// 这里视频文件得要存在
	if my_util.IsFile(videoListAdd.PhysicalVideoFileFullPath) == false {

		c.JSON(http.StatusOK, backend.ReplyJobThings{
			Message: "physical video file not found",
		})
		return
	}

	videoType := common.Movie
	if videoListAdd.VideoType == 1 {
		videoType = common.Series
	}
	nowJob := TTaskqueue.NewOneJob(
		videoType, videoListAdd.PhysicalVideoFileFullPath, videoListAdd.TaskPriorityLevel,
		videoListAdd.MediaServerInsideVideoID,
	)
	bok, err := cb.cronHelper.DownloadQueue.Add(*nowJob)
	if err != nil {
		return
	}
	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyJobThings{
			JobID:   nowJob.Id,
			Message: "job is already in queue",
		})
	} else {
		c.JSON(http.StatusOK, backend.ReplyJobThings{
			JobID:   nowJob.Id,
			Message: "ok",
		})
	}
}

// GetJobStatusHandler 外部 API 接口获取任务的状态
func (cb ControllerBase) GetJobStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "GetJobStatusHandler", err)
	}()

	jobID := c.DefaultQuery("job_id", "")
	if jobID == "" {
		err = errors.New("job_id is empty")
		return
	}

	found, oneJob := cb.cronHelper.DownloadQueue.GetOneJobByID(jobID)
	if found == false {
		err = errors.New("GetOneJobByID failed, id=" + jobID)
		return
	}
}
