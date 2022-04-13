package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestTaskQueue_AddAndGetAndDel(t *testing.T) {

	defer func() {
		DelDb()
	}()
	DelDb()

	taskQueue := NewTaskQueue("testQueue", settings.NewSettings(), log_helper.GetLogger())
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue.NewOneJob(common.Movie, "", i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJobs, err := taskQueue.Get(task_queue.Waiting)
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
		DelDb()
	}()
	DelDb()

	taskQueue := NewTaskQueue("testQueue", settings.NewSettings(), log_helper.GetLogger())
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue.NewOneJob(common.Movie, "", i))
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
		DelDb()
	}()
	DelDb()

	taskQueue := NewTaskQueue("testQueue", settings.NewSettings(), log_helper.GetLogger())
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue.NewOneJob(common.Movie, "", i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJobs, err := taskQueue.Get(task_queue.Waiting)
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

		waitingJob.JobStatus = task_queue.Committed

		bok, err = taskQueue.Update(waitingJob)
		if err != nil {
			t.Fatal("TestTaskQueue.Update", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Update == false")
		}
	}

	bok, commitedJobs, err := taskQueue.Get(task_queue.Committed)
	if err != nil {
		t.Fatal("TestTaskQueue.Get", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Get == false")
	}

	if len(commitedJobs) != taskPriorityCount+1 {
		t.Fatal("len(commitedJobs) != taskPriorityCount")
	}
}

func TestTaskQueue_UpdateAdGetOneWaiting(t *testing.T) {

	defer func() {
		DelDb()
	}()
	DelDb()

	taskQueue := NewTaskQueue("testQueue", settings.NewSettings(), log_helper.GetLogger())
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue.NewOneJob(common.Movie, spew.Sprintf("%d", i), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJob, err := taskQueue.GetOneWaiting()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaiting", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaiting == false")
	}

	if waitingJob.TaskPriority != 0 {
		t.Fatal("waitingJob.TaskPriority != 0")
	}

	waitingJob.JobStatus = task_queue.Committed
	bok, err = taskQueue.Update(waitingJob)
	if err != nil {
		t.Fatal("TestTaskQueue.Update", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.Update == false")
	}

	bok, waitingJob, err = taskQueue.GetOneWaiting()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaiting", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaiting == false")
	}

	if waitingJob.TaskPriority != 1 {
		t.Fatal("waitingJob.TaskPriority != 0")
	}
}

func TestTaskQueue_UpdatePriority(t *testing.T) {

	defer func() {
		DelDb()
	}()
	DelDb()

	taskQueue := NewTaskQueue("testQueue", settings.NewSettings(), log_helper.GetLogger())
	for i := taskPriorityCount; i >= 0; i-- {
		bok, err := taskQueue.Add(*task_queue.NewOneJob(common.Movie, spew.Sprintf("%d", i), i))
		if err != nil {
			t.Fatal("TestTaskQueue.Add", err)
		}
		if bok == false {
			t.Fatal("TestTaskQueue.Add == false")
		}
	}

	bok, waitingJob, err := taskQueue.GetOneWaiting()
	if err != nil {
		t.Fatal("TestTaskQueue.GetOneWaiting", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetOneWaiting == false")
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

	bok, waitingJobs, err := taskQueue.GetTaskPriority(0, task_queue.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.GetTaskPriority", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetTaskPriority == false")
	}

	if len(waitingJobs) != 0 {
		t.Fatal("len(waitingJobs) != 0")
	}

	bok, waitingJobs, err = taskQueue.GetTaskPriority(1, task_queue.Waiting)
	if err != nil {
		t.Fatal("TestTaskQueue.GetTaskPriority", err)
	}
	if bok == false {
		t.Fatal("TestTaskQueue.GetTaskPriority == false")
	}

	if len(waitingJobs) != 2 {
		t.Fatal("len(waitingJobs) != 2")
	}
}
