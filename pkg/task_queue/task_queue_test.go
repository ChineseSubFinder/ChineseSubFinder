package task_queue

import (
	"fmt"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	task_queue2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
)

const taskQueueName = "testQueue"

func TestTaskQueue_AddAndGetAndDel(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))
	defer func() {
		taskQueue.Close()
	}()
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, pkg.RandStringBytesMaskImprSrcSB(10), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJobs, err := taskQueue.GetJobsByStatus(task_queue2.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.Get", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Get == false")
	}

	if len(waitingJobs) != taskPriorityCount+1 {
		t.Fatal("len(waitingJobs) != taskPriorityCount")
	}

	for i := 0; i <= taskPriorityCount; i++ {

		if waitingJobs[i].TaskPriority != i {
			t.Fatalf("TestTaskQueue.TaskPriority pop error, want = %d, got = %d", i, waitingJobs[i].TaskPriority)
		}
	}

	for _, waitingJob := range waitingJobs {
		bok, err = taskQueue.Del(waitingJob.Id)
		if err != nil {
			t.Fatal("TestTaskQueue.Del", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Del == false")
		}
	}

	if taskQueue.Size() != 0 {
		t.Fatal("taskQueue.Size() != 0")
	}
}

func TestTaskQueue_AddAndClear(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, pkg.RandStringBytesMaskImprSrcSB(10), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	err := taskQueue.Clear()
	if err != nil {
		t.Fatal("TestTaskQueue.Clear", err)
	}

	if taskQueue.Size() != 0 {
		t.Fatal("taskQueue.Size() != 0")
	}
}

func TestTaskQueue_Update(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, pkg.RandStringBytesMaskImprSrcSB(10), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJobs, err := taskQueue.GetJobsByStatus(task_queue2.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.Get", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Get == false")
	}

	if len(waitingJobs) != taskPriorityCount+1 {
		t.Fatal("len(waitingJobs) != taskPriorityCount")
	}

	for i := 0; i <= taskPriorityCount; i++ {

		if waitingJobs[i].TaskPriority != i {
			t.Fatalf("TestTaskQueue.TaskPriority pop error, want = %d, got = %d", i, waitingJobs[i].TaskPriority)
		}
	}

	for _, waitingJob := range waitingJobs {

		waitingJob.JobStatus = task_queue2.Committed

		bok, err = taskQueue.Update(waitingJob)
		if err != nil {
			t.Fatal("TestTaskQueue.Update", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Update == false")
		}
	}

	bok, committedJobs, err := taskQueue.GetJobsByStatus(task_queue2.Committed)
	if err != nil {
		t.Fatal("TestTaskQueue.Get", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Get == false")
	}

	if len(committedJobs) != taskPriorityCount+1 {
		t.Fatal("len(committedJobs) != taskPriorityCount")
	}
}

func TestTaskQueue_UpdateAdGetOneWaiting(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, fmt.Sprintf("%d", i), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJob, err := taskQueue.GetOneWaitingJob()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaitingJob", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaitingJob == false")
	}

	if waitingJob.TaskPriority != 0 {
		t.Fatal("waitingJob.TaskPriority != 0")
	}

	waitingJob.JobStatus = task_queue2.Committed
	bok, err = taskQueue.Update(waitingJob)
	if err != nil {
		t.Fatal("TestTaskQueue.Update", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Update == false")
	}

	bok, waitingJob, err = taskQueue.GetOneWaitingJob()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaitingJob", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaitingJob == false")
	}

	if waitingJob.TaskPriority != 1 {
		t.Fatal("waitingJob.TaskPriority != 0")
	}
}

func TestTaskQueue_UpdatePriority(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, fmt.Sprintf("%d", i), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJob, err := taskQueue.GetOneWaitingJob()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaitingJob", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaitingJob == false")
	}

	if waitingJob.TaskPriority != 0 {
		t.Fatal("waitingJob.TaskPriority != 0")
	}

	waitingJob.TaskPriority = 1
	bok, err = taskQueue.Update(waitingJob)
	if err != nil {
		t.Fatal("TestTaskQueue.Update", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Update == false")
	}

	bok, waitingJobs, err := taskQueue.GetJobsByPriorityAndStatus(0, task_queue2.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.GetJobsByPriorityAndStatus", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetJobsByPriorityAndStatus == false")
	}

	if len(waitingJobs) != 0 {
		t.Fatal("len(waitingJobs) != 0")
	}

	bok, waitingJobs, err = taskQueue.GetJobsByPriorityAndStatus(1, task_queue2.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.GetJobsByPriorityAndStatus", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetJobsByPriorityAndStatus == false")
	}

	if len(waitingJobs) != 2 {
		t.Fatal("len(waitingJobs) != 2")
	}
}

func TestTaskQueue_AddAndGetOneJob(t *testing.T) {

	defer func() {
		cache_center.DelDb(taskQueueName)
	}()
	cache_center.DelDb(taskQueueName)

	taskQueue := NewTaskQueue(cache_center.NewCacheCenter(taskQueueName, log_helper.GetLogger4Tester()))

	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue2.NewOneJob(common.Movie, fmt.Sprintf("%d", i), DefaultTaskPriorityLevel))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, oneJob, err := taskQueue.GetOneJob()
	if err != nil {
		t.Fatal("TestTaskQueue.Add", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Add == false")
	}

	println("VideoFPath", oneJob.VideoFPath)
	println("TaskPriority", oneJob.TaskPriority)

	taskQueue.AutoDetectUpdateJobStatus(oneJob, nil)

	bok, oneJob, err = taskQueue.GetOneJob()
	if err != nil {
		t.Fatal("TestTaskQueue.Add", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Add == false")
	}

	println("VideoFPath", oneJob.VideoFPath)
	println("TaskPriority", oneJob.TaskPriority)

	found, waitingJobs, err := taskQueue.GetJobsByStatus(task_queue2.Waiting)
	if err != nil {
		return
	}
	println(found)
	for i, job := range waitingJobs {
		println("QueueDownloader Waiting:", i, job.VideoName)
	}

	found, waitingJobs, err = taskQueue.GetJobsByStatus(task_queue2.Done)
	if err != nil {
		return
	}
	println(found)
	for i, job := range waitingJobs {
		println("QueueDownloader Done:", i, job.VideoName)
	}

	found, waitingJobs, err = taskQueue.GetJobsByStatus(task_queue2.Failed)
	if err != nil {
		return
	}
	println(found)
	for i, job := range waitingJobs {
		println("QueueDownloader Failed:", i, job.VideoName)
	}

}
