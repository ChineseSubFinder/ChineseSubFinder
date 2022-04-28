package TestCode

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"time"
)

func baseHardWork(job interface{}) error {
	index := job.(int)
	println("Index:", index, "Start")
	time.Sleep(time.Second * 1)
	println("Index:", index, "End")
	return nil
}

func DoThings(ctx context.Context) error {
	const total = 10
	for i := 0; i < total; i++ {
		// 创建一个 chan 用于任务的中断和超时
		done := make(chan interface{}, 1)
		// 接收内部任务的 panic
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			// 匹配对应的 Eps 去处理
			done <- baseHardWork(i)
		}()

		select {
		case errInterface := <-done:
			if errInterface != nil {
				println("Index:", i, "Error:", errInterface)
			}
			break
		case p := <-panicChan:
			// 遇到内部的 panic，向外抛出
			panic(p)
		case <-ctx.Done():
			{
				err := errors.New(fmt.Sprintf("cancel at index: %d", i))
				return err
			}
		}
	}

	return nil
}

func MainProcess() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		err := DoThings(ctx)
		if err != nil {
			println("Error:", err.Error())
		}
	}()

	time.Sleep(5 * time.Second)
	cancel()
}
