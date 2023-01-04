package task_queue

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/emby"
	task_queue2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/sirupsen/logrus"
)

type TaskQueue struct {
	queueName           string                    // 队列的名称
	log                 *logrus.Logger            // 日志
	center              *cache_center.CacheCenter // 缓存中心
	taskPriorityMapList []*treemap.Map            // 这里有 0-10 个优先级划分的存储 List，每Add一个数据的时候需要切换到这个 List 中去 save
	taskKeyMap          *treemap.Map              // 以每个任务的唯一 JobID 来存储每个 Job 的 优先级在哪里，这样可以快速查询
	taskGroupBySeries   *treemap.Map              // 以每个任务的 SeriesRootPath 来存储每个任务，然后内层是一个 treeset，后续可以遍历删除即可
	queueLock           sync.Mutex                // 公用这个锁
}

func NewTaskQueue(center *cache_center.CacheCenter) *TaskQueue {

	tq := &TaskQueue{queueName: center.GetName(),
		log:                 center.Log,
		center:              center,
		taskPriorityMapList: make([]*treemap.Map, 0),
		taskKeyMap:          treemap.NewWithStringComparator(),
		taskGroupBySeries:   treemap.NewWithStringComparator(),
	}
	for i := 0; i <= taskPriorityCount; i++ {
		tq.taskPriorityMapList = append(tq.taskPriorityMapList, treemap.NewWithStringComparator())
	}
	tq.read()

	tq.afterRead()

	return tq
}

func (t *TaskQueue) Close() {
	t.center.Close()
}

func (t *TaskQueue) QueueName() string {
	return t.queueName
}

func (t *TaskQueue) Clear() error {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	err := t.center.TaskQueueClear()
	if err != nil {
		return err
	}

	for i := 0; i <= taskPriorityCount; i++ {
		t.taskPriorityMapList[i].Clear()
	}

	t.taskKeyMap.Clear()

	t.taskGroupBySeries.Clear()

	return nil
}

// Size 队列的长度，对外暴露，有锁
func (t *TaskQueue) Size() int {
	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	return t.taskKeyMap.Size()
}

// checkPriority 检测优先级，会校验范围
func (t *TaskQueue) checkPriority(oneJob task_queue2.OneJob) task_queue2.OneJob {

	if oneJob.TaskPriority > taskPriorityCount {
		oneJob.TaskPriority = taskPriorityCount
	}

	if oneJob.TaskPriority < 0 {
		oneJob.TaskPriority = 0
	}

	return oneJob
}

// degrade 降一级，会校验范围
func (t *TaskQueue) degrade(oneJob task_queue2.OneJob) task_queue2.OneJob {

	oneJob.TaskPriority -= 1

	return t.checkPriority(oneJob)
}

// Add 放入元素，放入的时候会根据 TaskPriority 进行归类，存在的不会新增和更新
func (t *TaskQueue) Add(oneJob task_queue2.OneJob) (bool, error) {

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
	// 如果是连续剧，则需要存储到 taskGroupBySeries 中
	jobIDSet, found := t.taskGroupBySeries.Get(oneJob.SeriesRootDirPath)
	if found == false {
		// 不存在
		nowJobIDSet := treeset.NewWithStringComparator()
		nowJobIDSet.Add(oneJob.Id)
		t.taskGroupBySeries.Put(oneJob.SeriesRootDirPath, nowJobIDSet)
	} else {
		// 存在
		nowJobIDSet := jobIDSet.(*treeset.Set)
		nowJobIDSet.Add(oneJob.Id)
		t.taskGroupBySeries.Put(oneJob.SeriesRootDirPath, nowJobIDSet)
	}
	err := t.save(oneJob.TaskPriority)
	if err != nil {
		return false, err
	}

	return true, nil
}

// update 更新素，不存在则会失败，内部用，没有锁
func (t *TaskQueue) update(oneJob task_queue2.OneJob) (bool, error) {

	if t.isExist(oneJob.Id) == false {
		return false, nil
	}
	// 自动更新时间
	oneJob.UpdateTime = (emby.Time)(time.Now())

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
func (t *TaskQueue) Update(oneJob task_queue2.OneJob) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	return t.update(oneJob)
}

// AutoDetectUpdateJobStatus 根据任务的生命周期图，进行自动判断更新，见《任务的生命周期》流程图
func (t *TaskQueue) AutoDetectUpdateJobStatus(oneJob task_queue2.OneJob, inErr error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	// 检查权限范围
	oneJob = t.checkPriority(oneJob)

	if inErr == nil {

		// 如果任务的优先级是 0，那么这个任务就认为是一次性任务，下载完毕不管如何都会设置为 ignore
		if oneJob.TaskPriority == 0 {
			oneJob.JobStatus = task_queue2.Ignore
		}

		// 没有错误就是完成
		oneJob.TaskPriority = DefaultTaskPriorityLevel
		oneJob.JobStatus = task_queue2.Done
		oneJob.DownloadTimes += 1
	} else {
		// 超过了时间限制，默认是 90 天, A.Before(B) : A < B == true
		if (time.Time)(oneJob.AddedTime).AddDate(0, 0, settings.Get().AdvancedSettings.TaskQueue.ExpirationTime).Before(time.Now()) == true {
			// 超过 90 天了
			oneJob.JobStatus = task_queue2.Failed
		} else {
			// 还在 90 天内
			// 是否是首次，那么就看它的 Level 是否是在 5，然后 retry == 0
			if oneJob.TaskPriority == DefaultTaskPriorityLevel && oneJob.RetryTimes == 0 {
				// 需要重置到 L6
				oneJob.RetryTimes = 0
				oneJob.TaskPriority = FirstRetryTaskPriorityLevel
			} else {
				if oneJob.RetryTimes > settings.Get().AdvancedSettings.TaskQueue.MaxRetryTimes {
					// 超过重试次数会进行一次降级，然后重置这个次数
					oneJob.RetryTimes = 0
					oneJob = t.degrade(oneJob)
				}
			}

			// 强制为 waiting
			oneJob.JobStatus = task_queue2.Waiting
		}

		// 如果任务的优先级是 0，那么这个任务就认为是一次性任务，下载完毕不管如何都会设置为 ignore
		if oneJob.TaskPriority == 0 {
			oneJob.JobStatus = task_queue2.Ignore
		}
		// 传入的错误需要放进来
		oneJob.ErrorInfo = inErr.Error()
		oneJob.DownloadTimes += 1
	}

	// 只要是进入完成标记流程的任务，如果优先级还是很高，那么就需要重置到默认优先级上
	if oneJob.TaskPriority < DefaultTaskPriorityLevel {
		oneJob.TaskPriority = DefaultTaskPriorityLevel
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

func (t *TaskQueue) del(jobId string) (bool, error) {
	if t.isExist(jobId) == false {
		return false, nil
	}

	taskPriority, bok := t.taskKeyMap.Get(jobId)
	if bok == false {
		return false, nil
	}
	// 删除连续剧的 tree.Map 里面的 tree.Set 的元素
	needDelJobObj, bok := t.taskPriorityMapList[taskPriority.(int)].Get(jobId)
	if bok == false {
		return false, nil
	}
	needDelJob := needDelJobObj.(task_queue2.OneJob)
	jobSetsObj, bok := t.taskGroupBySeries.Get(needDelJob.SeriesRootDirPath)
	if bok == false {
		return false, nil
	}
	jobSets := jobSetsObj.(*treeset.Set)
	jobSets.Remove(jobId)
	// 删除任务
	t.taskKeyMap.Remove(jobId)
	t.taskPriorityMapList[taskPriority.(int)].Remove(jobId)

	err := t.save(taskPriority.(int))
	if err != nil {
		return false, err
	}
	// 删除任务的时候也需要删除对应的日志
	pathRoot := filepath.Join(pkg.ConfigRootDirFPath(), "Logs")
	fileFPath := filepath.Join(pathRoot, common.OnceLogPrefix+jobId+".log")
	if pkg.IsFile(fileFPath) == true {
		err = os.Remove(fileFPath)
		if err != nil {
			t.log.Errorln("del job", jobId, "logfile,error:", err)
		}
	}

	return true, nil
}

// Del 删除一个元素
func (t *TaskQueue) Del(jobId string) (bool, error) {

	defer t.queueLock.Unlock()
	t.queueLock.Lock()

	return t.del(jobId)
}

func (t *TaskQueue) read() {

	taskQueueRead, err := t.center.TaskQueueRead()
	if err != nil {
		t.log.Errorln("read task queue TaskQueueRead error:", err)
		return
	}

	for i := 0; i <= taskPriorityCount; i++ {

		value, bok := taskQueueRead[i]
		if bok == false {
			continue
		}
		err = t.taskPriorityMapList[i].FromJSON(value)
		if err != nil {
			t.log.Errorln("read task queue FromJSON error:", err)
		}
		// 上面的操作仅仅是把 OneJob 的 JSON 弄了出来，还需要转换为 OneJob 的结构体
		// JobID - OneJob
		t.taskPriorityMapList[i].Each(func(key interface{}, value interface{}) {

			jsonString, err := json.Marshal(value)
			if err != nil {
				t.log.Panicln(err)
			}
			nowOneJob := task_queue2.OneJob{}
			err = json.Unmarshal(jsonString, &nowOneJob)
			if err != nil {
				t.log.Panicln(err)
			}
			t.taskPriorityMapList[i].Put(key, nowOneJob)
		})
		// 需要把几个优先级的map中的key汇总
		// JobID - OneJob
		t.taskPriorityMapList[i].Each(func(key interface{}, value interface{}) {
			// JobID -- taskPriority
			t.taskKeyMap.Put(key, i)
			// SeriesRootDirPath -- tree.Set(JobID)
			oneJob := value.(task_queue2.OneJob)
			jobIDSet, found := t.taskGroupBySeries.Get(oneJob.SeriesRootDirPath)
			if found == false {
				// 不存在
				nowJobIDSet := treeset.NewWithStringComparator()
				nowJobIDSet.Add(oneJob.Id)
				t.taskGroupBySeries.Put(oneJob.SeriesRootDirPath, nowJobIDSet)
			} else {
				// 存在
				nowJobIDSet := jobIDSet.(*treeset.Set)
				nowJobIDSet.Add(oneJob.Id)
				t.taskGroupBySeries.Put(oneJob.SeriesRootDirPath, nowJobIDSet)
			}
		})
	}
}

func (t *TaskQueue) afterRead() {
	// 将 downloading 的任务重置为 waiting
	for TaskPriority := 0; TaskPriority <= taskPriorityCount; TaskPriority++ {
		t.taskPriorityMapList[TaskPriority].Each(func(key interface{}, value interface{}) {

			nowOneJob := value.(task_queue2.OneJob)
			if nowOneJob.JobStatus == task_queue2.Downloading {
				nowOneJob.JobStatus = task_queue2.Waiting
				nowOneJob.DownloadTimes += 1
				bok, err := t.update(nowOneJob)
				if err != nil {
					t.log.Errorln("afterRead.update failed", err)
					return
				}
				if bok == false {
					t.log.Errorln("afterRead.update failed")
					return
				}
			}
		})
	}
}

// save 需要把改变的数据保持到 K/V 数据库中，这个没有锁，所以需要在 Sync 中使用，不对外开放
func (t *TaskQueue) save(taskPriority int) error {

	b, err := t.taskPriorityMapList[taskPriority].ToJSON()
	if err != nil {
		return err
	}

	err = t.center.TaskQueueSave(taskPriority, b)
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
	HighTaskPriorityLevel       = 3
	DefaultTaskPriorityLevel    = 5
	FirstRetryTaskPriorityLevel = 6
	LowTaskPriorityLevel        = 7
)

var (
	ErrNoSubFound = errors.New("No Sub Found")
)
