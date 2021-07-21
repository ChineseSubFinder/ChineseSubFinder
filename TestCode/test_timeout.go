package TestCode

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/go-rod/rod/lib/proto"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/net/context"
	"sync"
	"time"
)

func DownloadTest() error {

	testFunc := func(i interface{}) error {
		inData := i.(InputData)
		defer func() {
			println(inData.Index, "testFunc done.")
		}()
		println(inData.Index, "start...")

		err2 := oneStep(inData)
		if err2 != nil {
			return err2
		}

		return nil

		//return goStep(inData)
	}

	antPool, err := ants.NewPoolWithFunc(2, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		data.Ctx = ctx
		done := make(chan error, 1)
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			done <- testFunc(data)
		}()

		select {
		case err := <-done:
			if err != nil {
				println("done with Error", err.Error())
			}
			return
		case p := <-panicChan:
			println("got panic", p)
		case <-ctx.Done():
			println("got time out", ctx.Err())
			return
		}
	})
	if err != nil {
		return err
	}
	defer antPool.Release()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		err = antPool.Invoke(InputData{Index: i, Wg: &wg})
		if err != nil {
			println("antPool.Invoke", err)
		}
	}
	wg.Wait()

	println("All Done.")

	return nil
}

func goStep(inData InputData) error {
	outDataChan := make(chan int)
	for i := 0; i < 2; i++ {
		go func(cxt context.Context, in int) {

			var outData int
			outData = -1
			defer func() {
				println(inData.Index, in, "go func done")
				outDataChan <- outData
			}()

			browser, err := rod_helper.NewBrowser("")
			if err != nil {
				println(inData.Index, in, "rod_helper.NewBrowser", err)
				return
			}
			defer func() {
				browser.Close()
				println(inData.Index, in, "browser closed")
			}()

			ontTime := false

			for {
				select {
				case <-cxt.Done():
					return
				default:
					if ontTime == true {
						return
					}
					ontTime = true

					page, err := rod_helper.NewPageNavigate(browser, "https://www.baidu.com", 5*time.Second, 5)
					if err != nil {
						println("NewPageNavigate time out", err)
						return
					}
					page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
						UserAgent: pkg.RandomUserAgent(true),
					})
					err = page.WaitLoad()
					time.Sleep(10 * time.Second)

					outData = in
				}
			}

		}(inData.Ctx, i)
	}

	countResult := 0
	for {
		select {
		case <-inData.Ctx.Done():
			// 超时退出
			return nil
		case v, ok := <-outDataChan:
			if ok == true {
				println(inData.Index, "outData ok", v)
			} else {
				println(inData.Index, "outData not ok", v)
			}
			countResult++
			// 跳出，收到够反馈了
			if countResult == 2 {
				return nil
			}
		}
	}
}

func oneStep(inData InputData) error {
	browser, err := rod_helper.NewBrowser("")
	if err != nil {
		println(inData.Index, "rod_helper.NewBrowser", err)
		return err
	}
	defer func() {
		browser.Close()
		println(inData.Index, "browser closed")
	}()
	page, err := rod_helper.NewPageNavigate(browser, "https://www.baidu.com", 10*time.Second, 5)
	if err != nil {
		return err
	}
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: pkg.RandomUserAgent(true),
	})
	err = page.WaitLoad()
	time.Sleep(10 * time.Second)
	return nil
}

type InputData struct {
	Ctx   context.Context
	Index int
	Wg    *sync.WaitGroup
}
