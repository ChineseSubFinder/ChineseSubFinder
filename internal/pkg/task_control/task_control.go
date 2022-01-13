package task_control

import (
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type TaskControl struct {
	pollName            string
	antPoolBase         *ants.PoolWithFunc
	wgBase              sync.WaitGroup
	log                 *logrus.Logger
	oneCtxTimeOutSecond int
	bHold               bool
	released            bool
	// 传入的 func
	ctxFunc func(ctx context.Context, inData interface{}) error
	// 输入结构锁
	inputDataMap     map[int64]*TaskData
	inputDataMapLock sync.Mutex
	// 结束锁
	cancelMap     map[int64]context.CancelFunc
	cancelMapLock sync.Mutex
	// 执行情况, 0 是成功，1 是未执行，2 是错误或者超时
	executeInfoMap     map[int64]TaskState
	executeInfoMapLock sync.Mutex

	commonLock sync.Mutex
}

func NewTaskControl(pollName string, size int, oneCtxTimeOutSecond int, log *logrus.Logger) (*TaskControl, error) {

	var err error
	tc := TaskControl{}
	tc.pollName = pollName
	tc.oneCtxTimeOutSecond = oneCtxTimeOutSecond
	tc.log = log
	tc.inputDataMap = make(map[int64]*TaskData, 0)
	tc.cancelMap = make(map[int64]context.CancelFunc, 0)
	tc.executeInfoMap = make(map[int64]TaskState, 0)
	tc.antPoolBase, err = ants.NewPoolWithFunc(size, func(inData interface{}) {
		tc.baseFuncHandler(inData)
	})
	if err != nil {
		return nil, err
	}
	tc.wgBase = sync.WaitGroup{}
	return &tc, nil
}

// SetCtxProcessFunc 设置后续需要用到的单个任务的 Func，注意，如果之前的任务没有完成，不应该再次调用函数。建议进行 Release 后，再次调用
func (tc *TaskControl) SetCtxProcessFunc(pf func(ctx context.Context, inData interface{}) error) {
	tc.ctxFunc = pf
}

// Invoke 向 SetCtxProcessFunc 设置的 Func 中提交数据处理
func (tc *TaskControl) Invoke(inData *TaskData) error {

	// 需要先记录有那些 ID 进来，然后再记录那些是完整执行的，以及出错执行的
	tc.setExecuteStatus(inData.Index, NoExecute)

	err := tc.antPoolBase.Invoke(inData)
	if err != nil {
		tc.setTaskDataStatus(inData, Error)
		tc.setExecuteStatus(inData.Index, Error)
		return err
	}

	tc.log.Debugln("Index:", inData.Index, "Invoke inputDataMap Lock()")
	tc.inputDataMapLock.Lock()
	tc.inputDataMap[inData.Index] = inData
	tc.inputDataMapLock.Unlock()
	tc.log.Debugln("Index:", inData.Index, "Invoke inputDataMap UnLock()")

	return nil
}

func (tc *TaskControl) baseFuncHandler(inData interface{}) {

	data := inData.(*TaskData)

	defer func() {
		tc.log.Debugln("Index:", data.Index, "baseFuncHandler wg.Done()")
		tc.wgBase.Done()
	}()

	// 实际执行的时候
	tc.wgBase.Add(1)
	tc.log.Debugln("Index:", data.Index, "baseFuncHandler wg.Add()")

	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tc.oneCtxTimeOutSecond)*time.Second)
	defer func() {

		// 那么对应的需要取消掉 map 中的记录
		tc.cancelMapLock.Lock()
		delete(tc.cancelMap, data.Index)
		tc.cancelMapLock.Unlock()
		cancel()
	}()

	// 如果已经执行 Release 则返回
	tc.commonLock.Lock()
	if tc.released == true {
		return
	}
	tc.commonLock.Unlock()

	// 记录 cancel
	tc.log.Debugln("Index:", data.Index, "baseFuncHandler cancelMapLock Lock()")
	tc.cancelMapLock.Lock()
	tc.cancelMap[data.Index] = cancel
	tc.cancelMapLock.Unlock()
	tc.log.Debugln("Index:", data.Index, "baseFuncHandler cancelMapLock UnLock()")

	done := make(chan error, 1)
	panicChan := make(chan interface{}, 1)
	go func(ctx context.Context) {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()

		done <- tc.ctxFunc(ctx, inData)
	}(ctx)

	select {
	case err := <-done:
		if err != nil {

			tc.setTaskDataStatus(data, Error)
			tc.setExecuteStatus(data.Index, Error)
			tc.log.Errorln("PollName:", tc.pollName, "Index:", data.Index, "NewPoolWithFunc done with Error", err.Error())
		} else {
			tc.setTaskDataStatus(data, Success)
			tc.setExecuteStatus(data.Index, Success)
		}
		return
	case p := <-panicChan:

		tc.setTaskDataStatus(data, Error)
		tc.setExecuteStatus(data.Index, Error)
		tc.log.Errorln("PollName:", tc.pollName, "Index:", data.Index, "NewPoolWithFunc got panic", p)
		return
	case <-ctx.Done():

		tc.setTaskDataStatus(data, Error)
		tc.setExecuteStatus(data.Index, Error)
		tc.log.Errorln("PollName:", tc.pollName, "Index:", data.Index, "NewPoolWithFunc got time out", ctx.Err())
		return
	}
}

// Hold 自身进行阻塞，如果你是使用 Web 服务器，那么应该无需使用该方法
func (tc *TaskControl) Hold() {
	tc.commonLock.Lock()
	tc.bHold = true
	tc.commonLock.Unlock()
	tc.wgBase.Add(1)
	tc.log.Debugln("Hold wg.Add()")
	tc.wgBase.Wait()
}

func (tc *TaskControl) Release() {

	tc.log.Debugln("Release Start")

	tc.commonLock.Lock()
	tc.released = true
	tc.commonLock.Unlock()

	tc.log.Debugln("Release.Release")
	tc.antPoolBase.Release()

	tc.log.Debugln("Release cancel() Start")
	// 统一 cancel cancel
	tc.cancelMapLock.Lock()
	for i, cancelFunc := range tc.cancelMap {
		tc.log.Debugln("Release cancel() Index:", i)
		cancelFunc()
	}
	tc.cancelMapLock.Unlock()

	tc.log.Debugln("Release cancel() End")

	var bHold bool
	tc.commonLock.Lock()
	bHold = tc.bHold
	tc.commonLock.Unlock()
	if bHold == true {
		tc.log.Debugln("Release Hold wg.Done()")
		tc.wgBase.Done()
	}

	tc.log.Debugln("Release End")
}

func (tc *TaskControl) Reboot() {

	var release bool
	tc.commonLock.Lock()
	release = tc.released
	tc.commonLock.Unlock()

	if release == true {
		// 如果被释放了，那么第一次 Invoke 的时候需要重启这个 pool
		tc.antPoolBase.Reboot()
		// 需要把缓存的 map 清理掉
		tc.inputDataMapLock.Lock()
		tc.inputDataMap = make(map[int64]*TaskData, 0)
		tc.inputDataMapLock.Unlock()

		tc.cancelMapLock.Lock()
		tc.cancelMap = make(map[int64]context.CancelFunc, 0)
		tc.cancelMapLock.Unlock()

		tc.executeInfoMapLock.Lock()
		tc.executeInfoMap = make(map[int64]TaskState, 0)
		tc.executeInfoMapLock.Unlock()

		tc.commonLock.Lock()
		tc.released = false
		tc.commonLock.Unlock()
	}
}

// GetExecuteInfo 获取 所有 Invoke 的执行情况，需要在 下一次 Invoke 拿走，否则会清空
// 成功执行的、未执行的、执行错误（超时）的
func (tc *TaskControl) GetExecuteInfo() ([]int64, []int64, []int64) {

	successList := make([]int64, 0)
	noExecuteList := make([]int64, 0)
	errorList := make([]int64, 0)

	tc.executeInfoMapLock.Lock()

	for i, state := range tc.executeInfoMap {
		if state == Success {
			successList = append(successList, i)
		} else if state == NoExecute {
			noExecuteList = append(noExecuteList, i)
		} else if state == Error {
			errorList = append(errorList, i)
		}
	}

	tc.executeInfoMapLock.Unlock()

	return successList, noExecuteList, errorList
}

// GetResult 获取 TaskData 的反馈值，需要在 下一次 Invoke 拿走，否则会清空
func (tc *TaskControl) GetResult(index int64) (bool, *TaskData) {
	tc.inputDataMapLock.Lock()
	value, found := tc.inputDataMap[index]
	tc.inputDataMapLock.Unlock()
	return found, value
}

func (tc *TaskControl) setExecuteStatus(index int64, status TaskState) {
	tc.executeInfoMapLock.Lock()
	tc.executeInfoMap[index] = status
	tc.executeInfoMapLock.Unlock()
}

func (tc *TaskControl) setTaskDataStatus(taskData *TaskData, status TaskState) {
	tc.inputDataMapLock.Lock()
	taskData.Status = status
	tc.inputDataMapLock.Unlock()
}

type TaskData struct {
	Index            int64
	Status           TaskState // 执行情况, 0 是成功，1 是未执行，2 是错误或者超时
	OneVideoFullPath string
}

type TaskState int

const (
	Success TaskState = iota
	NoExecute
	Error
)
