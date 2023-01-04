package v1

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	task_queue3 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	task_queue2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_queue"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) JobsListHandler(c *gin.Context) {
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
		c.JSON(http.StatusOK, backend2.ReplyAllJobs{
			AllJobs: make([]task_queue3.OneJob, 0),
		})
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyAllJobs{
		AllJobs: allJobs,
	})
}

func (cb *ControllerBase) ChangeJobStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobsListHandler", err)
	}()

	desJobStatus := backend2.ReqChangeJobStatus{}
	err = c.ShouldBindJSON(&desJobStatus)
	if err != nil {
		return
	}

	bok, nowOneJob := cb.cronHelper.DownloadQueue.GetOneJobByID(desJobStatus.Id)
	if bok == false {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "job not found"})
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
	// 默认只能把任务改变为这两种状态
	if desJobStatus.JobStatus == task_queue3.Waiting || desJobStatus.JobStatus == task_queue3.Ignore {
		nowOneJob.JobStatus = desJobStatus.JobStatus
	} else {
		nowOneJob.JobStatus = task_queue3.Waiting
	}

	bok, err = cb.cronHelper.DownloadQueue.Update(nowOneJob)
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "update job status failed"})
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
}

func (cb *ControllerBase) JobLogHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobLogHandler", err)
	}()

	reqJobLog := backend2.ReqJobLog{}
	err = c.ShouldBindJSON(&reqJobLog)
	if err != nil {
		return
	}

	pathRoot := filepath.Join(pkg.ConfigRootDirFPath(), "Logs")
	fileFPath := filepath.Join(pathRoot, common.OnceLogPrefix+reqJobLog.Id+".log")
	if pkg.IsFile(fileFPath) == true {
		// 存在
		// 一行一行的读取文件
		var fi *os.File
		fi, err = os.Open(fileFPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		defer fi.Close()

		ReplyJobLog := backend2.ReplyJobLog{}
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
		return
	} else {
		// 不存在
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "job log not found"})
		return
	}
}
