package v1

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	TTaskqueue "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/gin-gonic/gin"
)

// AddJobHandler 外部 API 接口添加任务的处理
func (cb *ControllerBase) AddJobHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "AddJobHandler", err)
	}()

	videoListAdd := backend2.ReqVideoListAdd{}
	err = c.ShouldBindJSON(&videoListAdd)
	if err != nil {
		return
	}

	if videoListAdd.IsBluray == false {
		// 非蓝光的才需要检测这个文件存在
		// 这里视频文件得要存在
		if pkg.IsFile(videoListAdd.PhysicalVideoFileFullPath) == false {

			c.JSON(http.StatusOK, backend2.ReplyJobThings{
				Message: "physical video file not found",
			})
			return
		}
	}
	videoType := common.Movie
	if videoListAdd.VideoType == 1 {
		videoType = common.Series
	}
	nowJob := TTaskqueue.NewOneJob(
		videoType, videoListAdd.PhysicalVideoFileFullPath, videoListAdd.TaskPriorityLevel,
		videoListAdd.MediaServerInsideVideoID,
	)

	if videoListAdd.VideoType == 1 {
		// 连续剧
		// 连续剧的时候需要额外提交信息
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(videoListAdd.PhysicalVideoFileFullPath)
		if err != nil {
			return
		}
		seriesInfoDirPath := decode.GetSeriesDirRootFPath(videoListAdd.PhysicalVideoFileFullPath)
		if seriesInfoDirPath == "" {
			err = errors.New(fmt.Sprintf("decode.GetSeriesDirRootFPath == Empty, %s", videoListAdd.PhysicalVideoFileFullPath))
			return
		}
		nowJob.Season = epsVideoNfoInfo.Season
		nowJob.Episode = epsVideoNfoInfo.Episode
		nowJob.SeriesRootDirPath = seriesInfoDirPath
	}

	bok, err := cb.cronHelper.DownloadQueue.Add(*nowJob)
	if err != nil {
		return
	}
	if bok == false {
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			JobID:   nowJob.Id,
			Message: "job is already in queue",
		})
	} else {
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			JobID:   nowJob.Id,
			Message: "ok",
		})
	}
}

// GetJobStatusHandler 外部 API 接口获取任务的状态
func (cb *ControllerBase) GetJobStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "GetJobStatusHandler", err)
	}()

	jobID := c.DefaultQuery("job_id", "")
	if jobID == "" {
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "job_id is empty",
		})
		return
	}

	found, nowOneJob := cb.cronHelper.DownloadQueue.GetOneJobByID(jobID)
	if found == false {
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			JobID:   jobID,
			Message: "job not found",
		})
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyJobThings{
		JobID:     jobID,
		JobStatus: nowOneJob.JobStatus,
		Message:   "ok",
	})
}

// AddVideoPlayedInfoHandler 外部 API 接口添加已观看视频的信息
func (cb *ControllerBase) AddVideoPlayedInfoHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "AddVideoPlayedInfoHandler", err)
	}()

	videoPlayedInfo := backend2.ReqVideoPlayedInfo{}
	err = c.ShouldBindJSON(&videoPlayedInfo)
	if err != nil {
		return
	}
	// 这里视频文件得要存在
	if pkg.IsFile(videoPlayedInfo.PhysicalVideoFileFullPath) == false {

		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "physical video file not found",
		})
		return
	}
	// 查询字幕是否存在
	videoDirFPath := filepath.Dir(videoPlayedInfo.PhysicalVideoFileFullPath)
	subFileFullPath := filepath.Join(videoDirFPath, videoPlayedInfo.SubName)
	if pkg.IsFile(subFileFullPath) == false {

		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "sub file not found",
		})
		return
	}

	var videoPlayedInfos []models.ThirdPartSetVideoPlayedInfo
	dao.GetDb().Where("physical_video_file_full_path = ?", videoPlayedInfo.PhysicalVideoFileFullPath).Find(&videoPlayedInfos)
	if len(videoPlayedInfos) == 0 {
		// 没有则新增
		nowVideoPlayedInfo := models.ThirdPartSetVideoPlayedInfo{
			PhysicalVideoFileFullPath: videoPlayedInfo.PhysicalVideoFileFullPath,
			SubName:                   videoPlayedInfo.SubName,
		}
		dao.GetDb().Create(&nowVideoPlayedInfo)
	} else {
		// 有则更新
		videoPlayedInfos[0].SubName = videoPlayedInfo.SubName
		dao.GetDb().Save(&videoPlayedInfos[0])
	}

	c.JSON(http.StatusOK, backend2.ReplyJobThings{
		Message: "ok",
	})
}

// DelVideoPlayedInfoHandler 外部 API 接口删除已观看视频的信息
func (cb *ControllerBase) DelVideoPlayedInfoHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "DelVideoPlayedInfoHandler", err)
	}()

	videoPlayedInfo := backend2.ReqVideoPlayedInfo{}
	err = c.ShouldBindJSON(&videoPlayedInfo)
	if err != nil {
		return
	}
	// 这里视频文件得要存在
	if pkg.IsFile(videoPlayedInfo.PhysicalVideoFileFullPath) == false {

		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "physical video file not found",
		})
		return
	}

	var videoPlayedInfos []models.ThirdPartSetVideoPlayedInfo
	dao.GetDb().Where("physical_video_file_full_path = ?", videoPlayedInfo.PhysicalVideoFileFullPath).Find(&videoPlayedInfos)
	if len(videoPlayedInfos) == 0 {
		// 没有则也返回成功
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "ok",
		})
		return

	} else {
		// 有则更新，因为这个物理路径是主键，所以不用担心会查询出多个
		dao.GetDb().Delete(&videoPlayedInfos[0])
		c.JSON(http.StatusOK, backend2.ReplyJobThings{
			Message: "ok",
		})
		return
	}
}
