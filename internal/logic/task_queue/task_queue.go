package task_queue

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
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

// Add 放入元素，放入的时候会根据 TaskPriority 进行归类，存在的不会新增和更新
func (t *TaskQueue) Add(oneJob task_queue.OneJob) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	if t.isExist(oneJob.Id) == true {
		return false, nil
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

	if t.isExist(oneJob.Id) == false {
		return false, nil
	}
	// 自动更新时间
	oneJob.UpdateTime = time.Now()

	// 这里需要判断是否有优先级的 Update，如果有就需要把之前缓存的表给更新
	// 然后再插入到新的表中
	taskPriorityIndex, _ := t.taskKeyMap.Get(oneJob.Id)
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

// GetOneWaiting 获取一个元素，按优先级，0 - taskPriorityCount 的级别去拿去任务，不会移除任务
func (t *TaskQueue) GetOneWaiting() (bool, task_queue.OneJob, error) {

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
			if tOneJob.JobStatus == task_queue.Waiting {
				// 找到就返回
				found = true
				return
			}
		})

		if found == true {
			break
		}
	}

	return true, tOneJob, nil
}

func (t *TaskQueue) Get(status task_queue.JobStatus) (bool, []task_queue.OneJob, error) {

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

func (t *TaskQueue) GetTaskPriority(taskPriority int, status task_queue.JobStatus) (bool, []task_queue.OneJob, error) {

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
					if IsErrOk(err) == true {
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

// isExist 是否已经存在
func (t *TaskQueue) isExist(jobID string) bool {
	_, bok := t.taskKeyMap.Get(jobID)
	return bok
}

func (t *TaskQueue) isEmpty() bool {
	return t.taskKeyMap.Empty()
}

const taskPriorityCount = 10
