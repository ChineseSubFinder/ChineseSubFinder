package task_queue

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/badger_err_check"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	taskQueue2 "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/dgraph-io/badger/v3"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type TaskQueue struct {
	queueName           string
	settings            *settings.Settings
	log                 *logrus.Logger
	taskPriorityMapList []*treemap.Map
	taskKeyMap          *treemap.Map
	queueLock           sync.Mutex // 公用这个锁
}

func NewTaskQueue(queueName string, settings *settings.Settings, log *logrus.Logger) *TaskQueue {

	tq := &TaskQueue{queueName: queueName, settings: settings, log: log,
		taskPriorityMapList: make([]*treemap.Map, 0),
		taskKeyMap:          treemap.NewWithStringComparator(),
	}
	for i := 0; i <= taskPriorityCount; i++ {
		tq.taskPriorityMapList = append(tq.taskPriorityMapList, treemap.NewWithStringComparator())
	}
	tq.read()
	return tq
}

func (t *TaskQueue) QueueName() string {
	return t.queueName
}

func (t *TaskQueue) Clear() error {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	err := GetDb().Update(
		func(tx *badger.Txn) error {
			var err error

			for i := 0; i <= taskPriorityCount; i++ {
				key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue,
					fmt.Sprintf("%s_%d", t.queueName, i)))
				// 因为已经查询了一次，确保一定存在，所以直接更新+1，TTL 多加 5s 确保今天过去，暂时去除 TTL uint32(restOfDaySecond.Seconds())+5
				if err = tx.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		return err
	}

	for i := 0; i <= taskPriorityCount; i++ {
		t.taskPriorityMapList[i].Clear()
	}

	t.taskKeyMap.Clear()

	return nil
}

// Size 队列的长度，对外暴露，有锁
func (t *TaskQueue) Size() int {
	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	return t.taskKeyMap.Size()
}

// checkPriority 检测优先级，会校验范围
func (t *TaskQueue) checkPriority(oneJob taskQueue2.OneJob) taskQueue2.OneJob {

	if oneJob.TaskPriority > taskPriorityCount {
		oneJob.TaskPriority = taskPriorityCount
	}

	if oneJob.TaskPriority < 0 {
		oneJob.TaskPriority = 0
	}

	return oneJob
}

// degrade 降一级，会校验范围
func (t *TaskQueue) degrade(oneJob taskQueue2.OneJob) taskQueue2.OneJob {

	oneJob.TaskPriority -= 1

	return t.checkPriority(oneJob)
}

// Add 放入元素，放入的时候会根据 TaskPriority 进行归类，存在的不会新增和更新
func (t *TaskQueue) Add(oneJob task_queue.OneJob) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	if t.isExist(oneJob.Id) == true {
		return false, nil
	}
	// 检查权限范围
	oneJob = t.checkPriority(oneJob)
	// 插入到统一的 KeyMap
	t.taskKeyMap.Put(oneJob.Id, oneJob.TaskPriority)
	// 分配到具体的优先级 map 中
	t.taskPriorityMapList[oneJob.TaskPriority].Put(oneJob.Id, oneJob)
	err := t.save(oneJob.TaskPriority)
	if err != nil {
		return false, err
	}

	return true, nil
}

// update 更新素，不存在则会失败，内部用，没有锁
func (t *TaskQueue) update(oneJob task_queue.OneJob) (bool, error) {

	if t.isExist(oneJob.Id) == false {
		return false, nil
	}
	// 自动更新时间
	oneJob.UpdateTime = time.Now()

	// 这里需要判断是否有优先级的 Update，如果有就需要把之前缓存的表给更新
	// 然后再插入到新的表中
	taskPriorityIndex, _ := t.taskKeyMap.Get(oneJob.Id)
	// 检查权限范围
	oneJob = t.checkPriority(oneJob)
	if oneJob.TaskPriority != taskPriorityIndex {
		// 优先级修改
		// 先删除原有的优先级
		t.taskPriorityMapList[taskPriorityIndex.(int)].Remove(oneJob.Id)
		err := t.save(taskPriorityIndex.(int))
		if err != nil {
			return false, err
		}
	}
	// 插入到统一的 KeyMap
	t.taskKeyMap.Put(oneJob.Id, oneJob.TaskPriority)
	// 分配到具体的优先级 map 中
	t.taskPriorityMapList[oneJob.TaskPriority].Put(oneJob.Id, oneJob)
	err := t.save(oneJob.TaskPriority)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Update 更新素，不存在则会失败
func (t *TaskQueue) Update(oneJob task_queue.OneJob) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	return t.update(oneJob)
}

// AutoDetectUpdateJobStatus 根据任务的生命周期图，进行自动判断更新，见《任务的生命周期》流程图
func (t *TaskQueue) AutoDetectUpdateJobStatus(oneJob task_queue.OneJob, inErr error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	// 检查权限范围
	oneJob = t.checkPriority(oneJob)

	if inErr == nil {
		// 没有错误就是完成
		oneJob.TaskPriority = DefaultTaskPriorityLevel
		oneJob.JobStatus = taskQueue2.Done
		oneJob.DownloadTimes += 1
	} else {
		// 超过了时间限制，默认是 90 天, A.Before(B) : A < B == true
		if oneJob.AddedTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.ExpirationTime).Before(time.Now()) == true {
			// 超过 90 天了
			oneJob.JobStatus = taskQueue2.Failed
		} else {
			// 还在 90 天内
			// 是否是首次，那么就看它的 Level 是否是在 5，然后 retry == 0
			if oneJob.TaskPriority == DefaultTaskPriorityLevel && oneJob.RetryTimes == 0 {
				// 需要重置到 L6
				oneJob.RetryTimes = 0
				oneJob.TaskPriority = FirstRetryTaskPriorityLevel
			} else {
				if oneJob.RetryTimes > t.settings.AdvancedSettings.TaskQueue.MaxRetryTimes {
					// 超过重试次数会进行一次降级，然后重置这个次数
					oneJob.RetryTimes = 0
					oneJob = t.degrade(oneJob)
				}
			}

			// 强制为 waiting
			oneJob.JobStatus = taskQueue2.Waiting
		}
		// 传入的错误需要放进来
		oneJob.ErrorInfo = inErr.Error()
		oneJob.DownloadTimes += 1
	}

	// 这里不要用错了，要用无锁的，不然会阻塞
	bok, err := t.update(oneJob)
	if err != nil {
		t.log.Errorln("AutoDetectUpdateJobStatus", oneJob.VideoFPath, err)
		return
	}
	if bok == false {
		t.log.Warningln("AutoDetectUpdateJobStatus ==", oneJob.VideoFPath, "Job.ID", oneJob.Id, "Not Found")
		return
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

	found := false
	tOneJob := task_queue.OneJob{}
	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {

		t.taskPriorityMapList[TaskPriority].Each(func(key interface{}, value interface{}) {

			tOneJob = value.(task_queue.OneJob)
			// 任务的 UpdateTime 与现在的时间大于单个字幕下载的间隔
			// 默认是 12h, A.After(B) : A > B == true
			// 见《任务队列设计》--以优先级顺序取出描述
			if tOneJob.JobStatus == task_queue.Waiting && (tOneJob.DownloadTimes == 0 ||
				// 优先级 <= 3 也可以提前取出
				TaskPriority <= HightTaskPriorityLevel ||
				// 默认是 12h, A.After(B) : A > B == true
				tOneJob.UpdateTime.AddDate(0, 0, t.settings.AdvancedSettings.TaskQueue.OneSubDownloadInterval).After(time.Now()) == false && tOneJob.DownloadTimes > 0) {
				// 找到就返回
				found = true
				return
			}
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

		t.taskPriorityMapList[TaskPriority].Each(func(key interface{}, value interface{}) {

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
				return
			}
		})

		if found == true {
			return true, tOneJob, nil
		}
	}

	return false, tOneJob, nil
}

func (t *TaskQueue) GetJobsByStatus(status task_queue.JobStatus) (bool, []task_queue.OneJob, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	outOneJobs := make([]task_queue.OneJob, 0)
	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, nil, nil
	}

	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {

		t.taskPriorityMapList[TaskPriority].Each(func(key interface{}, value interface{}) {

			tOneJob := task_queue.OneJob{}
			tOneJob = value.(task_queue.OneJob)
			if tOneJob.JobStatus == status {
				// 找到加入列表
				outOneJobs = append(outOneJobs, tOneJob)
			}
		})
	}

	return true, outOneJobs, nil
}

// GetJobsByPriorityAndStatus 根据任务优先级和状态获取任务列表
func (t *TaskQueue) GetJobsByPriorityAndStatus(taskPriority int, status task_queue.JobStatus) (bool, []task_queue.OneJob, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	outOneJobs := make([]task_queue.OneJob, 0)
	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, nil, nil
	}

	t.taskPriorityMapList[taskPriority].Each(func(key interface{}, value interface{}) {

		tOneJob := task_queue.OneJob{}
		tOneJob = value.(task_queue.OneJob)
		if tOneJob.JobStatus == status {
			// 找到加入列表
			outOneJobs = append(outOneJobs, tOneJob)
		}
	})

	return true, outOneJobs, nil
}

// Del 删除一个元素
func (t *TaskQueue) Del(jobId string) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	if t.isExist(jobId) == false {
		return false, nil
	}

	taskPriority, bok := t.taskKeyMap.Get(jobId)
	if bok == false {
		return false, nil
	}
	t.taskKeyMap.Remove(jobId)
	t.taskPriorityMapList[taskPriority.(int)].Remove(jobId)

	err := t.save(taskPriority.(int))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (t *TaskQueue) read() {

	err := GetDb().View(
		func(tx *badger.Txn) error {
			var err error
			for i := 0; i <= taskPriorityCount; i++ {

				key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue,
					fmt.Sprintf("%s_%d", t.queueName, i)))
				var item *badger.Item
				item, err = tx.Get(key)
				if err != nil {
					if badger_err_check.IsErrOk(err) == true {
						return nil
					}
					return err
				}
				valCopy, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				err = t.taskPriorityMapList[i].FromJSON(valCopy)
				if err != nil {
					return err
				}
			}

			return nil
		})
	if err != nil {
		t.log.Panicln(err)
	}
	// 需要把几个优先级的map中的key汇总
	for i := 0; i < taskPriorityCount; i++ {
		t.taskPriorityMapList[i].Each(func(key interface{}, value interface{}) {
			t.taskKeyMap.Put(key, i)
		})
	}
}

// save 需要把改变的数据保持到 K/V 数据库中，这个没有锁，所以需要在 Sync 中使用，不对外开放
func (t *TaskQueue) save(taskPriority int) error {

	err := GetDb().Update(
		func(tx *badger.Txn) error {
			var err error

			key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue,
				fmt.Sprintf("%s_%d", t.queueName, taskPriority)))
			if err != nil {
				return err
			}

			b, err := t.taskPriorityMapList[taskPriority].ToJSON()
			if err != nil {
				return err
			}
			e := badger.NewEntry(key, b)
			err = tx.SetEntry(e)
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		return err
	}

	return nil
}

// isExist 是否已经存在，对内，无锁
func (t *TaskQueue) isExist(jobID string) bool {
	_, bok := t.taskKeyMap.Get(jobID)
	return bok
}

// IsExist 是否已经存在，对外，有锁
func (t *TaskQueue) IsExist(jobID string) bool {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	_, bok := t.taskKeyMap.Get(jobID)
	return bok
}

// isEmpty 对内，无锁
func (t *TaskQueue) isEmpty() bool {
	return t.taskKeyMap.Empty()
}

const (
	taskPriorityCount           = 10
	HightTaskPriorityLevel      = 3
	DefaultTaskPriorityLevel    = 5
	FirstRetryTaskPriorityLevel = 6
)

var (
	ErrNotSubFound = errors.New("Not Sub Found")
)
