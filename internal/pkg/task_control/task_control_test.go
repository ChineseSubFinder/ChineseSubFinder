package task_control

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestNewTaskControl(t *testing.T) {

	// 不超时的情况
	err := process(TimeTester{
		ConcurrentCount:  2,
		JobCount:         5,
		TimeAfterRelease: 5,
		OneJobWaitTime:   2,
		OneJobTimeOut:    5,
		SelfHold:         true,
		DontRelease:      false,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = process(TimeTester{
		ConcurrentCount:  2,
		JobCount:         5,
		TimeAfterRelease: 5,
		OneJobWaitTime:   2,
		OneJobTimeOut:    5,
		SelfHold:         false,
		DontRelease:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = process(TimeTester{
		ConcurrentCount:  2,
		JobCount:         5,
		TimeAfterRelease: 5,
		OneJobWaitTime:   2,
		OneJobTimeOut:    5,
		SelfHold:         false,
		DontRelease:      false,
	})
	if err != nil {
		t.Fatal(err)
	}

}

func process(timeTester TimeTester) error {

	OneJobWaitTime = timeTester.OneJobWaitTime

	tc, err := NewTaskControl("TestPool", timeTester.ConcurrentCount, timeTester.OneJobTimeOut, log_helper.GetLogger())
	if err != nil {
		return err
	}
	tc.SetCtxProcessFunc(waitTimes)

	for i := 0; i < timeTester.JobCount; i++ {
		go func(index int) {
			if index > 1 {
				time.Sleep(10 * time.Second)
			}
			err := tc.Invoke(InputData{Index: index})
			if err != nil {
				fmt.Println("Index:", index, "Error", err)
			}
		}(i)
	}

	go func() {
		if timeTester.DontRelease == true {
			fmt.Println("dont Release")
			return
		}
		fmt.Println("Release After 2 Second")
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

	return nil
}

func waitTimes(ctx context.Context, inData interface{}) error {

	phase0 := make(chan interface{}, 1)

	index := inData.(InputData)

	fmt.Println("Index:", index.Index, "Start 0")
	time.Sleep(time.Duration(OneJobWaitTime) * time.Second)
	fmt.Println("Index:", index.Index, "End 0")

	phase0 <- 1
	select {
	case <-ctx.Done():
		{
			fmt.Println("Index:", index.Index, "timeout 0")
			return nil
		}
	case <-phase0:
		break
	}

	fmt.Println("Index:", index.Index, "Start 1")
	fmt.Println("Index:", index.Index, "End 1")

	return nil
}

type TimeTester struct {
	ConcurrentCount  int  // 并发数
	JobCount         int  // 总任务数
	TimeAfterRelease int  // 开始后等待多久执行 Release 操作
	OneJobWaitTime   int  // 单个任务得耗时
	OneJobTimeOut    int  // 单个任务的超时时间
	SelfHold         bool // 是否需要自身的等待，如果使用了，那么一定需要 Release
	DontRelease      bool // 是否需要主动执行 Release
}

var OneJobWaitTime int
