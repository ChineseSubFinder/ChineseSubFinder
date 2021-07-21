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

		browser, err := rod_helper.NewBrowser("")
		if err != nil {
			return err
		}
		defer func() {
			browser.Close()
			println(inData.Index, "browser closed")
		}()

		page, err := rod_helper.NewPageNavigate(browser, "https://www.baidu.com", 1*time.Second, 5)
		if err != nil {
			return err
		}
		page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: pkg.RandomUserAgent(true),
		})
		err = page.WaitLoad()
		time.Sleep(6 * time.Second)

		return nil
	}

	antPool, err := ants.NewPoolWithFunc(2, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan error, 1)
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			done <- testFunc(inData)
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

type InputData struct {
	Index int
	Wg    *sync.WaitGroup
}
