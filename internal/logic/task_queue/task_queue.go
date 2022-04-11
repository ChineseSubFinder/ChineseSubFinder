package task_queue

import (
	"encoding/json"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/dgraph-io/badger/v3"
	dll "github.com/emirpasic/gods/lists/doublylinkedlist"
	"github.com/sirupsen/logrus"
	"sync"
)

type TaskQueue struct {
	queueName      string
	settings       *settings.Settings
	log            *logrus.Logger
	doubleList     *dll.List
	doubleListLock sync.Mutex
	lockList       bool
	lockListLock   sync.Mutex
}

func NewTaskQueue(queueName string, settings *settings.Settings, log *logrus.Logger) *TaskQueue {

	tq := &TaskQueue{queueName: queueName, settings: settings, log: log, doubleList: dll.New()}
	tq.read()
	return tq
}

func (t *TaskQueue) QueueName() string {
	return t.queueName
}

func (t *TaskQueue) Clear() error {

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	t.doubleList.Clear()

	err := GetDb().Update(
		func(tx *badger.Txn) error {
			var err error
			key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue, t.queueName))
			// 因为已经查询了一次，确保一定存在，所以直接更新+1，TTL 多加 5s 确保今天过去，暂时去除 TTL uint32(restOfDaySecond.Seconds())+5
			if err = tx.Delete(key); err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return err
	}

	return nil
}

// isEmpty 这个没有锁，所以需要在 Sync 中使用，不对外开放
func (t *TaskQueue) isEmpty() bool {

	if t.doubleList.Size() > 0 {
		return false
	}

	return true
}

// IsEmpty 是否队列为空，对外暴露，有锁
func (t *TaskQueue) IsEmpty() bool {
	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	return t.isEmpty()
}

func (t *TaskQueue) Size() int {
	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	return t.doubleList.Size()
}

func (t *TaskQueue) read() {

	err := GetDb().View(
		func(tx *badger.Txn) error {
			var err error

			key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue, t.queueName))
			e, err := tx.Get(key)
			if err != nil {

				if IsErrOk(err) == true {
					return nil
				}

				return err
			}
			valCopy, err := e.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(valCopy, t.doubleList)
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		t.log.Panicln(err)
	}
}

// save 需要把改变的数据保持到 K/V 数据库中，这个没有锁，所以需要在 Sync 中使用，不对外开放
func (t *TaskQueue) save() error {

	err := GetDb().Update(
		func(tx *badger.Txn) error {
			var err error
			key := []byte(MergeBucketAndKeyName(BucketNamePrefixVideoSubDownloadQueue, t.queueName))
			if err != nil {
				return err
			}
			b, err := json.Marshal(t.doubleList)
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

// RPush 向右边放入元素
func (t *TaskQueue) RPush(oneJob task_queue.OneJob) (bool, error) {

	if t.isLockList() == true {
		return false, nil
	}

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	t.doubleList.Add(oneJob)

	err := t.save()
	if err != nil {
		return false, err
	}

	return true, nil
}

// LPush 向左边放入元素
func (t *TaskQueue) LPush(oneJob task_queue.OneJob) (bool, error) {

	if t.isLockList() == true {
		return false, nil
	}

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	t.doubleList.Add(0, oneJob)

	err := t.save()
	if err != nil {
		return false, err
	}

	return true, nil
}

// RPop 从右边取出第一个元素，并移除
func (t *TaskQueue) RPop() (bool, task_queue.OneJob, error) {

	if t.isLockList() == true {
		return false, task_queue.OneJob{}, nil
	}

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}, nil
	}

	rightLastOneIndex := t.doubleList.Size() - 1
	value, bok := t.doubleList.Get(rightLastOneIndex)
	if bok == false {
		return false, task_queue.OneJob{}, nil
	}
	// 移除最后一个元素
	t.doubleList.Remove(rightLastOneIndex)

	err := t.save()
	if err != nil {
		return false, task_queue.OneJob{}, err
	}

	return true, value.(task_queue.OneJob), nil
}

// RPeek 获取右边取出第一个元素，不移除
func (t *TaskQueue) RPeek() (bool, task_queue.OneJob) {

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}
	}

	rightLastOneIndex := t.doubleList.Size() - 1
	value, bok := t.doubleList.Get(rightLastOneIndex)
	if bok == false {
		return false, task_queue.OneJob{}
	}

	return true, value.(task_queue.OneJob)
}

// LPop 向左边取出第一个元素，并移除
func (t *TaskQueue) LPop() (bool, task_queue.OneJob, error) {

	if t.isLockList() == true {
		return false, task_queue.OneJob{}, nil
	}

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}, nil
	}

	leftFistOneIndex := 0
	value, bok := t.doubleList.Get(leftFistOneIndex)
	if bok == false {
		return false, task_queue.OneJob{}, nil
	}
	// 移除左边第一个元素
	t.doubleList.Remove(leftFistOneIndex)

	err := t.save()
	if err != nil {
		return false, task_queue.OneJob{}, err
	}

	return true, value.(task_queue.OneJob), nil
}

// LPeek 向左边获取第一个元素，不移除
func (t *TaskQueue) LPeek() (bool, task_queue.OneJob) {

	defer t.doubleListLock.Unlock()
	t.doubleListLock.Lock()

	// 如果队列里面没有东西，则返回 false
	if t.isEmpty() == true {
		return false, task_queue.OneJob{}
	}

	leftFistOneIndex := 0
	value, bok := t.doubleList.Get(leftFistOneIndex)
	if bok == false {
		return false, task_queue.OneJob{}
	}

	return true, value.(task_queue.OneJob)
}

// LockList 锁住 List，这样才能够正确的进行遍历
func (t *TaskQueue) LockList() {
	defer t.lockListLock.Unlock()
	t.lockListLock.Lock()

	t.lockList = true
}

// UnLockList 解锁 List，就可以正常的 Push 和 Pop
func (t *TaskQueue) UnLockList() {
	defer t.lockListLock.Unlock()
	t.lockListLock.Lock()

	t.lockList = false
}

func (t *TaskQueue) isLockList() bool {

	bLock := false
	t.lockListLock.Lock()
	bLock = t.lockList
	t.lockListLock.Unlock()

	return bLock
}

// GetList 使用的时候不要插入数据，否则会有问题
func (t *TaskQueue) GetList() dll.Iterator {
	return t.doubleList.Iterator()
}
