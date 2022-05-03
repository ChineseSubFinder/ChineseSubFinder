package v1

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	task_queue2 "github.com/allanpk716/ChineseSubFinder/internal/pkg/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (cb ControllerBase) JobsListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobsListHandler", err)
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

func (cb ControllerBase) ChangeJobStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobsListHandler", err)
	}()

	desJobStatus := backend.ReqChangeJobStatus{}
	err = c.ShouldBindJSON(&desJobStatus)
	if err != nil {
		return
	}

	bok, nowOneJob := cb.cronHelper.DownloadQueue.GetOneJobByID(desJobStatus.Id)
	if bok == false {
		err = errors.New("")
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "job not found"})
		return
	}

	if desJobStatus.TaskPriority == "high" {
		// high
		nowOneJob.TaskPriority = task_queue2.HighTaskPriorityLevel
	} else if desJobStatus.TaskPriority == "mddile" {
		// middle
		nowOneJob.TaskPriority = task_queue2.DefaultTaskPriorityLevel
	} else {
		// low
		nowOneJob.TaskPriority = task_queue2.LowTaskPriorityLevel
	}
	nowOneJob.JobStatus = task_queue.Waiting

	bok, err = cb.cronHelper.DownloadQueue.Update(nowOneJob)
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "update job status failed"})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
}

func (cb ControllerBase) JobLogHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobLogHandler", err)
	}()

	reqJobLog := backend.ReqJobLog{}
	err = c.ShouldBindJSON(&reqJobLog)
	if err != nil {
		return
	}

	pathRoot := filepath.Join(global_value.ConfigRootDirFPath(), "Logs")
	fileFPath := filepath.Join(pathRoot, common.OnceLogPrefix+reqJobLog.Id+".log")
	if my_util.IsFile(fileFPath) == true {
		// 存在
		// 一行一行的读取文件
		var fi *os.File
		fi, err = os.Open(fileFPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		defer fi.Close()

		ReplyJobLog := backend.ReplyJobLog{}
		ReplyJobLog.OneLine = make([]string, 0)
		br := bufio.NewReader(fi)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			ReplyJobLog.OneLine = append(ReplyJobLog.OneLine, string(a))
		}

		c.JSON(http.StatusOK, ReplyJobLog)
	} else {
		// 不存在
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "job log not found"})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
}
