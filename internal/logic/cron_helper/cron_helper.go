package cron_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/downloader_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_download_process"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

type CronHelper struct {
	fullSubDownloadProcessing     bool // 这个是核心耗时函数执行的状态
	fullSubDownloadProcessingLock sync.Mutex
	cronHelperRunning             bool // 这个是定时器启动的状态，它为true，不代表核心函数在执行
	cronHelperRunningLock         sync.Mutex
	c                             *cron.Cron
	dh                            *downloader_helper.DownloaderHelper
}

func NewCronHelper() (*CronHelper, error) {

	ch := CronHelper{}
	ch.c = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := ch.c.AddFunc("@every "+settings.GetSettings().CommonSettings.ScanInterval, ch.coreSubDownloadProcess)
	if err != nil {
		log_helper.GetLogger().Errorln("CronHelper Cron entryID:", entryID, "Error:", err)
		return nil, err
	}
	return &ch, nil
}

// Start 开启定时器任务，这个任务是非阻塞的，coreSubDownloadProcess 仅仅可能是这个函数执行耗时而已
// runImmediately == false 那么 ch.c.Start() 是不会阻塞的
func (ch *CronHelper) Start(runImmediately bool) {

	ch.cronHelperRunningLock.Lock()
	ch.cronHelperRunning = true
	ch.cronHelperRunningLock.Unlock()
	// 是否在定时器开启前先执行一次任务
	if runImmediately == true {

		log_helper.GetLogger().Infoln("First Time coreSubDownloadProcess Start")

		ch.coreSubDownloadProcess()

		log_helper.GetLogger().Infoln("First Time coreSubDownloadProcess End")

	} else {
		log_helper.GetLogger().Infoln("RunAtStartup: false, so will not Run At Startup, wait",
			settings.GetSettings().CommonSettings.ScanInterval, "to Download")
	}

	log_helper.GetLogger().Infoln("CronHelper Start...")
	log_helper.GetLogger().Infoln("Next Sub Scan Will Process After", settings.GetSettings().CommonSettings.ScanInterval)
	ch.c.Start()

	// 只有定时任务 start 之后才能拿到信息
	if len(ch.c.Entries()) > 0 {
		// 不会马上启动扫描，那么就需要设置当前的时间，且为 waiting
		tttt := ch.c.Entries()[0].Next.Format("2006-01-02 15:04:05")
		common.SetSubScanJobStatusWaiting(tttt)
	} else {
		log_helper.GetLogger().Errorln("Can't get cron jobs, will not send SubScanJobStatus")
	}
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

	ch.cronHelperRunningLock.Lock()
	ch.cronHelperRunning = false
	ch.cronHelperRunningLock.Unlock()

	common.SetSubScanJobStatusNil()
}

func (ch *CronHelper) CronHelperRunning() bool {

	defer func() {
		ch.cronHelperRunningLock.Unlock()
	}()
	ch.cronHelperRunningLock.Lock()
	return ch.cronHelperRunning
}

func (ch *CronHelper) CronRunningStatusString() string {
	if ch.CronHelperRunning() == true {
		return Running
	} else {
		return Stopped
	}
}

func (ch *CronHelper) FullDownloadProcessRunning() bool {

	defer func() {
		ch.fullSubDownloadProcessingLock.Unlock()
	}()
	ch.fullSubDownloadProcessingLock.Lock()
	return ch.fullSubDownloadProcessing
}

func (ch *CronHelper) FullDownloadProcessRunningStatusString() string {
	if ch.FullDownloadProcessRunning() == true {
		return Running
	} else {
		return Stopped
	}
}

// coreSubDownloadProcess 执行一次下载任务的多个步骤
func (ch *CronHelper) coreSubDownloadProcess() {

	defer func() {
		ch.fullSubDownloadProcessingLock.Lock()
		ch.fullSubDownloadProcessing = false
		ch.fullSubDownloadProcessingLock.Unlock()

		log_helper.GetLogger().Infoln(log_helper.OnceSubsScanEnd)
	}()

	ch.fullSubDownloadProcessingLock.Lock()
	ch.fullSubDownloadProcessing = true
	ch.fullSubDownloadProcessingLock.Unlock()

	log_helper.GetLogger().Infoln(log_helper.OnceSubsScanStart)

	// 扫描字幕任务开始，先是扫描阶段，那么是拿不到有多少视频需要扫描的数量的
	common.SetSubScanJobStatusPreparing(time.Now().Format("2006-01-02 15:04:05"))

	// 下载前的初始化
	preDownloadProcess := pre_download_process.NewPreDownloadProcess()
	err := preDownloadProcess.
		Init().
		Check().
		HotFix().
		ChangeSubNameFormat().
		ReloadBrowser().
		Wait()
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

const (
	Stopped = "stopped"
	Running = "running"
)
