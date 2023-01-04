package task_control

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func TestTaskControl_Invoke(t *testing.T) {
	type args struct {
		timeTester TimeTester
	}
	tests := []struct {
		name                string
		args                args
		successProcessCount int
		wantErr             bool
	}{
		// 不超时的情况
		{
			name: "00", args: args{
				TimeTester{PoolName: "00",
					ConcurrentCount: 1,
					JobCount:        5,
					OneJobWaitTime:  1,
					OneJobTimeOut:   2,
				}},
			successProcessCount: 5,
		},
		{
			name: "01", args: args{
				TimeTester{PoolName: "01",
					ConcurrentCount: 2,
					JobCount:        5,
					OneJobWaitTime:  1,
					OneJobTimeOut:   2,
				}},
			successProcessCount: 5,
		},
		{
			name: "02", args: args{
				TimeTester{PoolName: "02",
					ConcurrentCount: 3,
					JobCount:        5,
					OneJobWaitTime:  1,
					OneJobTimeOut:   2,
				}},
			successProcessCount: 5,
		},
		// 超时的情况
		{
			name: "03", args: args{
				TimeTester{PoolName: "03",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,
					NeedRelease:      true}},
			successProcessCount: 0,
		},
		{
			name: "04", args: args{
				TimeTester{PoolName: "04",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,
					NeedRelease:      false}},
			successProcessCount: 0,
		},
		{
			name: "05", args: args{
				TimeTester{PoolName: "05",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,
					NeedRelease:      true}},
			successProcessCount: 0,
		},
		// 主动触发 painic
		{
			name: "06", args: args{
				TimeTester{PoolName: "06",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,

					NeedRelease: true,
					WantPanic:   true}},
			successProcessCount: 0,
		},
		{
			name: "07", args: args{
				TimeTester{PoolName: "07",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,

					NeedRelease: false,
					WantPanic:   true}},
			successProcessCount: 0,
		},
		{
			name: "08", args: args{
				TimeTester{PoolName: "08",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   2,
					OneJobTimeOut:    1,

					NeedRelease: true,
					WantPanic:   true}},
			successProcessCount: 0,
		},
		// 部分超时
		{
			name: "09", args: args{
				TimeTester{PoolName: "09",
					ConcurrentCount:          2,
					JobCount:                 5,
					TimeAfterRelease:         5,
					OneJobWaitTime:           2,
					OneJobTimeOut:            3,
					NeedRelease:              true,
					IndexOverThanAddMoreTime: 2}},
			successProcessCount: 3,
		},
		{
			name: "10", args: args{
				TimeTester{PoolName: "10",
					ConcurrentCount:          2,
					JobCount:                 5,
					TimeAfterRelease:         5,
					OneJobWaitTime:           2,
					OneJobTimeOut:            3,
					NeedRelease:              false,
					IndexOverThanAddMoreTime: 2}},
			successProcessCount: 3,
		},
		{
			name: "11", args: args{
				TimeTester{PoolName: "11",
					ConcurrentCount:          2,
					JobCount:                 5,
					TimeAfterRelease:         5,
					OneJobWaitTime:           2,
					OneJobTimeOut:            3,
					NeedRelease:              true,
					IndexOverThanAddMoreTime: 3}},
			successProcessCount: 4,
		},
		// 使用 Release 取消
		{
			name: "12", args: args{
				TimeTester{PoolName: "12",
					ConcurrentCount:  1,
					JobCount:         5,
					TimeAfterRelease: 2,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,
					NeedRelease:      true}},
			successProcessCount: 0,
		},
		{
			name: "13", args: args{
				TimeTester{PoolName: "13",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 2,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 0,
		},
		{
			name: "14", args: args{
				TimeTester{PoolName: "14",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 4,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 2,
		},
		{
			name: "15", args: args{
				TimeTester{PoolName: "15",
					ConcurrentCount:  1,
					JobCount:         5,
					TimeAfterRelease: 3,
					OneJobWaitTime:   2,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 1,
		},
		{
			name: "16", args: args{
				TimeTester{PoolName: "16",
					ConcurrentCount:  3,
					JobCount:         5,
					TimeAfterRelease: 4,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 3,
		},
		{
			name: "17", args: args{
				TimeTester{PoolName: "17",
					ConcurrentCount:  4,
					JobCount:         5,
					TimeAfterRelease: 4,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 4,
		},
		{
			name: "18", args: args{
				TimeTester{PoolName: "18",
					ConcurrentCount:  5,
					JobCount:         5,
					TimeAfterRelease: 4,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,

					NeedRelease: true}},
			successProcessCount: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successList, _, _, err := process(tt.name, tt.args.timeTester)
			if err != nil {
				t.Fatal(err)
			}

			if tt.successProcessCount != len(successList) {
				t.Fatal("want successProcessCount =", tt.successProcessCount, "now =", len(successList))
			}
		})
	}
}

func process(name string, timeTester TimeTester) ([]int, []int, []int, error) {

	once := sync.Once{}

	tc, err := NewTaskControl(timeTester.ConcurrentCount, log_helper.NewLogHelper(name, pkg.ConfigRootDirFPath(), logrus.DebugLevel, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour))
	if err != nil {
		return nil, nil, nil, err
	}
	tc.SetCtxProcessFunc(timeTester.PoolName, waitTimes, timeTester.OneJobTimeOut)

	for i := 0; i < timeTester.JobCount; i++ {

		once.Do(func() {
			go func() {
				if timeTester.NeedRelease == false {
					tc.log.Infoln("Do not need Release")
					return
				}
				tc.log.Infoln("Release After", timeTester.TimeAfterRelease, "Second")
				time.Sleep(time.Duration(timeTester.TimeAfterRelease) * time.Second)
				tc.Release()
			}()
		})

		err := tc.Invoke(&TaskData{Index: i,
			DataEx: DataEx{
				OneJobWaitTime:           timeTester.OneJobWaitTime,
				WantPanic:                timeTester.WantPanic,
				IndexOverThanAddMoreTime: timeTester.IndexOverThanAddMoreTime,
			}})
		if err != nil {
			tc.log.Errorln("Index:", i, "Error", err)
		}
	}

	tc.Hold()

	fmt.Println("-------------------------------")

	// 获取提前终止的计数器以及完成的计数器
	successList, noExecuteList, errorList := tc.GetExecuteInfo()

	return successList, noExecuteList, errorList, nil
}

func waitTimes(ctx context.Context, inData interface{}) error {

	phase0 := make(chan interface{}, 1)
	index := inData.(*TaskData)

	dataEx := index.DataEx.(DataEx)

	if dataEx.WantPanic == true {
		panic("want panic")
	}

	go func() {
		defer func() {
			close(phase0)
		}()
		fmt.Println("Index:", index.Index, "Start 0")
		if dataEx.IndexOverThanAddMoreTime == 0 {
			time.Sleep(time.Duration(dataEx.OneJobWaitTime) * time.Second)
		} else {
			if index.Index > dataEx.IndexOverThanAddMoreTime {
				time.Sleep(time.Duration(dataEx.OneJobWaitTime+10) * time.Second)
			} else {
				time.Sleep(time.Duration(dataEx.OneJobWaitTime) * time.Second)
			}
		}
		phase0 <- 1
		fmt.Println("Index:", index.Index, "End 0")
	}()

	select {
	case <-ctx.Done():
		{
			fmt.Println("Index:", index.Index, "timeout 0")
			return errors.New("timeout jump")
		}
	case <-phase0:
		break
	}

	fmt.Println("Index:", index.Index, "Start 1")
	fmt.Println("Index:", index.Index, "End 1")

	return nil
}

type TimeTester struct {
	PoolName                 string // 名称
	ConcurrentCount          int    // 并发数
	JobCount                 int    // 总任务数
	TimeAfterRelease         int    // 开始后等待多久执行 Release 操作
	OneJobWaitTime           int    // 单个任务得耗时
	OneJobTimeOut            int    // 单个任务的超时时间
	NeedRelease              bool   // 是否需要主动执行 Release
	WantPanic                bool   // 触发 panic
	IndexOverThanAddMoreTime int    // waitTimes函数中某个 Index 之后都会在等待处理上多加延时以便触发超时逻辑
}

type DataEx struct {
	OneJobWaitTime           int
	WantPanic                bool
	IndexOverThanAddMoreTime int
}
