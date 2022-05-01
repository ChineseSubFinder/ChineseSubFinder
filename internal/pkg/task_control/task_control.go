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
	released            bool
	// 传入的 func
	ctxFunc func(ctx context.Context, inData interface{}) error
	// 输入结构锁
	inputDataMap     map[int]*TaskData
	inputDataMapLock sync.Mutex
	// 结束锁
	cancelMap     map[int]context.CancelFunc
	cancelMapLock sync.Mutex
	// 执行情况, 0 是成功，1 是未执行，2 是错误或者超时
	executeInfoMap     map[int]TaskState
	executeInfoMapLock sync.Mutex

	commonLock sync.Mutex
}

func NewTaskControl(size int, log *logrus.Logger) (*TaskControl, error) {

	var err error
	tc := TaskControl{}
	tc.log = log
	tc.inputDataMap = make(map[int]*TaskData, 0)
	tc.cancelMap = make(map[int]context.CancelFunc, 0)
	tc.executeInfoMap = make(map[int]TaskState, 0)
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
func (tc *TaskControl) SetCtxProcessFunc(pollName string, pf func(ctx context.Context, inData interface{}) error, oneCtxTimeOutSecond int) {

	tc.pollName = pollName
	tc.ctxFunc = pf
	tc.oneCtxTimeOutSecond = oneCtxTimeOutSecond
}

// Invoke 向 SetCtxProcessFunc 设置的 Func 中提交数据处理
func (tc *TaskControl) Invoke(inData *TaskData) error {

	// 实际执行的时候
	tc.wgBase.Add(1)
	tc.log.Debugln("Index:", inData.Index, "baseFuncHandler wg.Add()")

	// 需要先记录有那些 ID 进来，然后再记录那些是完整执行的，以及出错执行的
	tc.setExecuteStatus(inData.Index, NoExecute)

	err := tc.antPoolBase.Invoke(inData)
	if err != nil {
		tc.setTaskDataStatus(inData, NoExecute)
		tc.setExecuteStatus(inData.Index, NoExecute)
		tc.log.Debugln("Index:", inData.Index, "baseFuncHandler wg.Done()")
		tc.wgBase.Done()
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

	// 如果已经执行 Release 则返回
	nowRelease := false
	tc.commonLock.Lock()
	nowRelease = tc.released
	tc.commonLock.Unlock()

	if nowRelease == true {
		tc.log.Debugln("Index:", data.Index, "released == true")
		return
	}

	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tc.oneCtxTimeOutSecond)*time.Second)
	defer func() {

		// 那么对应的需要取消掉 map 中的记录
		tc.log.Debugln("Index:", data.Index, "baseFuncHandler cancelMapLock Lock() defer")
		tc.cancelMapLock.Lock()
		delete(tc.cancelMap, data.Index)
		tc.cancelMapLock.Unlock()
		tc.log.Debugln("Index:", data.Index, "baseFuncHandler cancelMapLock UnLock() defer")
		cancel()

	}()

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
			close(done)
			close(panicChan)
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
	tc.log.Debugln("Hold()")
	tc.wgBase.Wait()
}

func (tc *TaskControl) Release() {

	tc.log.Debugln("-------------------------------")
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

	tc.log.Debugln("Release End")
	tc.log.Debugln("-------------------------------")
}

func (tc *TaskControl) Reboot() {

	tc.log.Debugln("-------------------------------")
	tc.log.Debugln("Reboot Start")
	var release bool
	tc.commonLock.Lock()
	release = tc.released
	tc.commonLock.Unlock()

	if release == true {
		// 如果被释放了，那么第一次 Invoke 的时候需要重启这个 pool
		tc.antPoolBase.Reboot()
		// 需要把缓存的 map 清理掉
		tc.inputDataMapLock.Lock()
		tc.inputDataMap = make(map[int]*TaskData, 0)
		tc.inputDataMapLock.Unlock()

		tc.cancelMapLock.Lock()
		tc.cancelMap = make(map[int]context.CancelFunc, 0)
		tc.cancelMapLock.Unlock()

		tc.executeInfoMapLock.Lock()
		tc.executeInfoMap = make(map[int]TaskState, 0)
		tc.executeInfoMapLock.Unlock()

		tc.commonLock.Lock()
		tc.released = false
		tc.commonLock.Unlock()
	}

	tc.log.Debugln("Reboot End")
	tc.log.Debugln("-------------------------------")
}

func (tc *TaskControl) Close() {
	tc.log.Debugln("-------------------------------")
	tc.log.Debugln("Close Start")
	tc.log.Debugln("Close End")
	tc.log.Debugln("-------------------------------")
}

// GetExecuteInfo 获取 所有 Invoke 的执行情况，需要在 下一次 Invoke 拿走，否则会清空
// 成功执行的、未执行的、执行错误（超时）的
func (tc *TaskControl) GetExecuteInfo() ([]int, []int, []int) {

	successList := make([]int, 0)
	noExecuteList := make([]int, 0)
	errorList := make([]int, 0)

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
func (tc *TaskControl) GetResult(index int) (bool, *TaskData) {
	tc.inputDataMapLock.Lock()
	value, found := tc.inputDataMap[index]
	tc.inputDataMapLock.Unlock()
	return found, value
}

func (tc *TaskControl) setExecuteStatus(index int, status TaskState) {
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
	Index  int         // 第几个任务
	Count  int         // 总任务的数量
	Status TaskState   // 执行情况, 0 是成功，1 是未执行，2 是错误或者超时
	DataEx interface{} // 需要传递到执行函数中的数据
}

type TaskState int

const (
	Success TaskState = iota
	NoExecute
	Error
)
