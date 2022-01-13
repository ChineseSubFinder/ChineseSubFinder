package task_control

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"golang.org/x/net/context"
	"testing"
	"time"
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
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   1,
					OneJobTimeOut:    2,
					SelfHold:         true,
					NeedRelease:      true}},
			successProcessCount: 5,
		},
		{
			name: "01", args: args{
				TimeTester{PoolName: "01",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   1,
					OneJobTimeOut:    2,
					SelfHold:         false,
					NeedRelease:      false}},
			successProcessCount: 5,
		},
		{
			name: "02", args: args{
				TimeTester{PoolName: "02",
					ConcurrentCount:  2,
					JobCount:         5,
					TimeAfterRelease: 5,
					OneJobWaitTime:   1,
					OneJobTimeOut:    2,
					SelfHold:         false,
					NeedRelease:      true}},
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
					SelfHold:         true,
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
					SelfHold:         false,
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
					SelfHold:         false,
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
					SelfHold:         true,
					NeedRelease:      true,
					WantPanic:        true}},
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
					SelfHold:         false,
					NeedRelease:      false,
					WantPanic:        true}},
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
					SelfHold:         false,
					NeedRelease:      true,
					WantPanic:        true}},
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
					SelfHold:                 true,
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
					SelfHold:                 false,
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
					SelfHold:                 false,
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
					SelfHold:         true,
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
					SelfHold:         true,
					NeedRelease:      true}},
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
					SelfHold:         true,
					NeedRelease:      true}},
			successProcessCount: 2,
		},
		{
			name: "15", args: args{
				TimeTester{PoolName: "15",
					ConcurrentCount:  1,
					JobCount:         5,
					TimeAfterRelease: 4,
					OneJobWaitTime:   3,
					OneJobTimeOut:    4,
					SelfHold:         true,
					NeedRelease:      true}},
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
					SelfHold:         true,
					NeedRelease:      true}},
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
					SelfHold:         true,
					NeedRelease:      true}},
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
					SelfHold:         true,
					NeedRelease:      true}},
			successProcessCount: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successList, _, _, err := process(tt.args.timeTester)
			if err != nil {
				t.Fatal(err)
			}

			if tt.successProcessCount != len(successList) {
				t.Fatal("want successProcessCount =", tt.successProcessCount, "now =", len(successList))
			}
		})
	}
}

func process(timeTester TimeTester) ([]int64, []int64, []int64, error) {

	OneJobWaitTime = timeTester.OneJobWaitTime
	WantPanic = timeTester.WantPanic
	IndexOverThanAddMoreTime = int64(timeTester.IndexOverThanAddMoreTime)

	tc, err := NewTaskControl(timeTester.PoolName, timeTester.ConcurrentCount, timeTester.OneJobTimeOut, log_helper.GetLogger())
	if err != nil {
		return nil, nil, nil, err
	}
	tc.SetCtxProcessFunc(waitTimes)

	for i := 0; i < timeTester.JobCount; i++ {
		go func(index int64) {
			err := tc.Invoke(&TaskData{Index: index})
			if err != nil {
				tc.log.Errorln("Index:", index, "Error", err)
			}
		}(int64(i))
	}

	go func() {
		if timeTester.NeedRelease == false {
			tc.log.Infoln("Do not need Release")
			return
		}
		tc.log.Infoln("Release After", timeTester.TimeAfterRelease, "Second")
		time.Sleep(time.Duration(timeTester.TimeAfterRelease) * time.Second)
		tc.Release()
	}()

	fmt.Println("-------------------------------")
	if timeTester.SelfHold == true {

		fmt.Println("Start Hold")
		tc.Hold()
		fmt.Println("End Hold")
	} else {

		waitTime := timeTester.JobCount * timeTester.OneJobWaitTime
		fmt.Printf("wait %ds start\n", waitTime)
		time.Sleep(time.Duration(waitTime) * time.Second)
		fmt.Printf("wait %ds end\n", waitTime)
	}
	fmt.Println("-------------------------------")

	// 获取提前终止的计数器以及完成的计数器
	successList, noExecuteList, errorList := tc.GetExecuteInfo()
	return successList, noExecuteList, errorList, nil
}

func waitTimes(ctx context.Context, inData interface{}) error {

	phase0 := make(chan interface{}, 1)
	index := inData.(*TaskData)
	if WantPanic == true {
		panic("want panic")
	}

	go func() {
		fmt.Println("Index:", index.Index, "Start 0")
		if IndexOverThanAddMoreTime == 0 {
			time.Sleep(time.Duration(OneJobWaitTime) * time.Second)
		} else {
			if index.Index > IndexOverThanAddMoreTime {
				time.Sleep(time.Duration(OneJobWaitTime+10) * time.Second)
			} else {
				time.Sleep(time.Duration(OneJobWaitTime) * time.Second)
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
	SelfHold                 bool   // 是否需要自身的等待，如果使用了，那么一定需要 Release
	NeedRelease              bool   // 是否需要主动执行 Release
	WantPanic                bool   // 触发 panic
	IndexOverThanAddMoreTime int    // waitTimes函数中某个 Index 之后都会在等待处理上多加延时以便触发超时逻辑
}

var OneJobWaitTime int
var WantPanic bool
var IndexOverThanAddMoreTime int64
