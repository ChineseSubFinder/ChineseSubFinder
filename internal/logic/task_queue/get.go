package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"time"
)

func (t *TaskQueue) BeforeGetOneJob() {
	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	// 这里需要手动判断 Done 的任务是否超过三个月了，超过就需要手动删除
	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {
		t.taskPriorityMapList[TaskPriority].Each(func(key interface{}, value interface{}) {

			nowOneJob := value.(task_queue.OneJob)
			if nowOneJob.JobStatus == task_queue.Done &&
				// 默认是 90day, A.After(B) : A > B == true
				nowOneJob.UpdateTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.ExpirationTime).After(time.Now()) == false {
				// 找到就删除
				bok, err := t.del(nowOneJob.Id)
				if err != nil {
					t.log.Errorf("GetOneWaitingJob.Del.Done ExpirationTime %v error: %s", t.settings.AdvancedSettings.TaskQueue.ExpirationTime, err.Error())
					return
				}
				if bok == false {
					t.log.Errorf("GetOneWaitingJob.Del.Done ExpirationTime %v error: %s", t.settings.AdvancedSettings.TaskQueue.ExpirationTime, "Del failed")
					return

				}
				return
			}
		})
	}
}

// GetOneJob 优先获取 GetOneWaitingJob 然后才是 GetOneDoneJob
func (t *TaskQueue) GetOneJob() (bool, task_queue.OneJob, error) {
	found, waitingJob, err := t.GetOneWaitingJob()
	if err != nil {
		return false, task_queue.OneJob{}, err
	}
	if found == false {
		return t.GetOneDoneJob()
	}

	return true, waitingJob, nil
}

// GetOneWaitingJob 获取一个元素，按优先级，0 - taskPriorityCount 的级别去拿去任务，不会移除任务
func (t *TaskQueue) GetOneWaitingJob() (bool, task_queue.OneJob, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}, nil
	}
	// 找到需要返回的复合条件的任务
	found := false
	tOneJob := task_queue.OneJob{}
	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {

		t.taskPriorityMapList[TaskPriority].Any(func(key interface{}, value interface{}) bool {

			tOneJob = value.(task_queue.OneJob)
			// 任务的 UpdateTime 与现在的时间大于单个字幕下载的间隔
			// 默认是 12h, A.After(B) : A > B == true
			// 见《任务队列设计》--以优先级顺序取出描述
			if tOneJob.JobStatus == task_queue.Waiting && (tOneJob.DownloadTimes == 0 ||
				// 优先级 <= 3 也可以提前取出
				TaskPriority <= HighTaskPriorityLevel ||
				// 默认是 12h, A.After(B) : A > B == true
				tOneJob.UpdateTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.OneSubDownloadInterval).After(time.Now()) == false && tOneJob.DownloadTimes > 0) {
				// 找到就返回
				found = true
				return true
			}

			return false
		})

		if found == true {
			return true, tOneJob, nil
		}
	}

	return false, tOneJob, nil
}

// GetOneDoneJob 获取一个元素，按优先级，0 - taskPriorityCount 的级别去拿去任务，不会移除任务
func (t *TaskQueue) GetOneDoneJob() (bool, task_queue.OneJob, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}, nil
	}

	found := false
	tOneJob := task_queue.OneJob{}
	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {

		t.taskPriorityMapList[TaskPriority].Any(func(key interface{}, value interface{}) bool {

			tOneJob = value.(task_queue.OneJob)
			// 任务的 UpdateTime 与现在的时间大于单个字幕下载的间隔
			// 默认是 12h, A.After(B) : A > B == true
			// 见《任务队列设计》--以优先级顺序取出描述
			if tOneJob.JobStatus == task_queue.Done &&
				// 要在 三个月内
				tOneJob.CreatedTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.ExpirationTime).After(time.Now()) == true &&
				// 已经下载过的视频，要间隔 12 小时再次下载
				tOneJob.UpdateTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.OneSubDownloadInterval).After(time.Now()) == false {
				// 找到就返回
				found = true
				return true
			}

			return false
		})

		if found == true {
			return true, tOneJob, nil
		}
	}

	return false, tOneJob, nil
}
