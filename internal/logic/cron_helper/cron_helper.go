package cron_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/downloader_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_download_process"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

type CronHelper struct {
	runImmediately                bool
	fullSubDownloadProcessing     bool
	fullSubDownloadProcessingLock sync.Locker
	cronHelperRunning             bool
	c                             *cron.Cron
	dh                            *downloader_helper.DownloaderHelper
}

func NewCronHelper() (*CronHelper, error) {

	ch := CronHelper{}
	ch.c = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := ch.c.AddFunc("@every "+settings.GetSettings().CommonSettings.ScanInterval, ch.fullSubDownloadProcess)
	if err != nil {
		log_helper.GetLogger().Errorln("CronHelper Cron entryID:", entryID, "Error:", err)
		return nil, err
	}
	return &ch, nil
}

// Start 开启定时器任务，这个任务是非阻塞的，fullSubDownloadProcess 仅仅可能是这个函数执行耗时而已
// runImmediately == false 那么 ch.c.Start() 是不会阻塞的
func (ch *CronHelper) Start(runImmediately bool) {

	ch.runImmediately = runImmediately
	// 是否在定时器开启前先执行一次任务
	if ch.runImmediately == true {

		log_helper.GetLogger().Infoln("First Time fullSubDownloadProcess Start")

		ch.fullSubDownloadProcess()

		log_helper.GetLogger().Infoln("First Time fullSubDownloadProcess End")

	} else {
		log_helper.GetLogger().Infoln("RunAtStartup: false, so will not Run At Startup, wait",
			settings.GetSettings().CommonSettings.ScanInterval, "to Download")
	}

	ch.c.Start()
}

// Stop 会阻塞等待任务完成
func (ch *CronHelper) Stop() {

	fullSubDownloadProcessing := false
	ch.fullSubDownloadProcessingLock.Lock()
	fullSubDownloadProcessing = ch.fullSubDownloadProcessing
	ch.fullSubDownloadProcessingLock.Unlock()

	if fullSubDownloadProcessing == true {
		if ch.dh != nil {
			ch.dh.Cancel()
		}
		// Stop stops the cron scheduler if it is running; otherwise it does nothing.
		// A context is returned so the caller can wait for running jobs to complete.
		nowContext := ch.c.Stop()
		select {
		case <-time.After(5 * time.Minute):
			log_helper.GetLogger().Warningln("Wait over 5 min, CronHelper is timeout")
		case <-nowContext.Done():
			log_helper.GetLogger().Infoln("CronHelper.Stop() Done.")
		}
	} else {
		// Stop stops the cron scheduler if it is running; otherwise it does nothing.
		// A context is returned so the caller can wait for running jobs to complete.
		nowContext := ch.c.Stop()
		select {
		case <-time.After(5 * time.Second):
			log_helper.GetLogger().Warningln("Wait over 5 s, CronHelper is timeout")
		case <-nowContext.Done():
			log_helper.GetLogger().Infoln("CronHelper.Stop() Done.")
		}
	}
}

func (ch *CronHelper) Running() bool {

	defer func() {
		ch.fullSubDownloadProcessingLock.Unlock()
	}()

	ch.fullSubDownloadProcessingLock.Lock()
	return ch.fullSubDownloadProcessing
}

// fullSubDownloadProcess 执行一次下载任务的多个步骤
func (ch *CronHelper) fullSubDownloadProcess() {

	defer func() {
		ch.fullSubDownloadProcessingLock.Lock()
		ch.fullSubDownloadProcessing = false
		ch.fullSubDownloadProcessingLock.Unlock()
	}()

	ch.fullSubDownloadProcessingLock.Lock()
	ch.fullSubDownloadProcessing = true
	ch.fullSubDownloadProcessingLock.Unlock()

	// 下载前的初始化
	preDownloadProcess := pre_download_process.NewPreDownloadProcess().
		Init().
		Check().
		HotFix().
		ChangeSubNameFormat().
		ReloadBrowser()
	err := preDownloadProcess.Wait()
	if err != nil {
		log_helper.GetLogger().Errorln("pre_download_process", "Error:", err)
		log_helper.GetLogger().Errorln("Skip DownloaderHelper.Start()")
		return
	}
	// 开始下载
	ch.dh = downloader_helper.NewDownloaderHelper(*settings.GetSettings(true),
		preDownloadProcess.SubSupplierHub)
	err = ch.dh.Start()
	if err != nil {
		log_helper.GetLogger().Errorln("downloader_helper.Start()", "Error:", err)
		return
	}
}
