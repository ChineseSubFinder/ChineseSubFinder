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
	ctxFunc             func(ctx context.Context, inData interface{}) error
}

func NewTaskControl(pollName string, size int, oneCtxTimeOutSecond int, log *logrus.Logger) (*TaskControl, error) {

	var err error
	tc := TaskControl{}
	tc.pollName = pollName
	tc.oneCtxTimeOutSecond = oneCtxTimeOutSecond
	tc.log = log
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
func (tc *TaskControl) Invoke(inData InputData) error {
	tc.wgBase.Add(1)
	inData.Wg = &tc.wgBase
	tc.log.Debugln("Index:", inData.Index, "Invoke wg.Add()")
	err := tc.antPoolBase.Invoke(inData)
	if err != nil {
		// 如果这个执行有问题，那么就把 wg 的计数器减一
		tc.log.Debugln("Index:", inData.Index, "Invoke Error wg.Done()")
		tc.wgBase.Done()
	}

	return err
}

func (tc *TaskControl) baseFuncHandler(inData interface{}) {
	data := inData.(InputData)
	defer func() {
		tc.log.Debugln("Index:", data.Index, "baseFuncHandler wg.Done()")
		data.Wg.Done()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tc.oneCtxTimeOutSecond)*time.Second)
	defer cancel()

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
			tc.log.Errorln(tc.pollName, "Index:", data.Index, "NewPoolWithFunc done with Error", err.Error())
		}
		return
	case p := <-panicChan:
		tc.log.Errorln(tc.pollName, "Index:", data.Index, "NewPoolWithFunc got panic", p)
		return
	case <-ctx.Done():
		tc.log.Errorln(tc.pollName, "Index:", data.Index, "NewPoolWithFunc got time out", ctx.Err())
		return
	}
}

// Hold 自身进行阻塞，如果你是使用 Web 服务器，那么应该无需使用该方法
func (tc *TaskControl) Hold() {
	tc.bHold = true
	tc.wgBase.Add(1)
	tc.log.Debugln("Hold wg.Add()")
	tc.wgBase.Wait()
}

func (tc *TaskControl) Release() {
	if tc.bHold == true {
		tc.log.Debugln("Release wg.Done()")
		tc.wgBase.Done()
	}
	tc.log.Debugln("Release.Release")
	tc.antPoolBase.Release()
}

type InputData struct {
	OneVideoFullPath string
	Index            int
	Wg               *sync.WaitGroup
}
