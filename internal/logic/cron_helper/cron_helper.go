package cron_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/downloader_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_download_process"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
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

	sets *settings.Settings
	log  *logrus.Logger
}

func NewCronHelper(_log *logrus.Logger, _sets *settings.Settings) *CronHelper {

	ch := CronHelper{
		log:  _log,
		sets: _sets,
	}
	return &ch
}

// Start 开启定时器任务，这个任务是非阻塞的，coreSubDownloadProcess 仅仅可能是这个函数执行耗时而已
// runImmediately == false 那么 ch.c.Start() 是不会阻塞的
func (ch *CronHelper) Start(runImmediately bool) {

	_, err := cron.ParseStandard(ch.sets.CommonSettings.ScanInterval)
	if err != nil {
		ch.log.Warningln("CommonSettings.ScanInterval format error, after v0.25.x , need reset this at WebUI")
		// 如果解析错误了，就需要重新赋值默认值过来，然后保存
		nowSettings := ch.sets
		nowSettings.CommonSettings.ScanInterval = settings.NewCommonSettings().ScanInterval
		err = settings.SetFullNewSettings(nowSettings)
		if err != nil {
			ch.log.Panicln("CronHelper.SetFullNewSettings:", err)
			return
		}
	}

	ch.c = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := ch.c.AddFunc(ch.sets.CommonSettings.ScanInterval, ch.coreSubDownloadProcess)
	if err != nil {
		ch.log.Panicln("CronHelper Cron entryID:", entryID, "Error:", err)
	}

	ch.cronHelperRunningLock.Lock()
	ch.cronHelperRunning = true
	ch.cronHelperRunningLock.Unlock()
	// 是否在定时器开启前先执行一次任务
	if runImmediately == true {

		ch.log.Infoln("First Time coreSubDownloadProcess Start")

		ch.coreSubDownloadProcess()

		ch.log.Infoln("First Time coreSubDownloadProcess End")

	} else {
		ch.log.Infoln("RunAtStartup: false, so will not Run At Startup")
	}

	ch.log.Infoln("CronHelper Start...")
	ch.c.Start()

	// 只有定时任务 start 之后才能拿到信息
	if len(ch.c.Entries()) > 0 {

		// 不会马上启动扫描，那么就需要设置当前的时间，且为 waiting
		tttt := ch.c.Entries()[0].Next.Format("2006-01-02 15:04:05")
		common.SetSubScanJobStatusWaiting(tttt)

		ch.log.Infoln("Next Sub Scan Will Process At:", tttt)
	} else {
		ch.log.Errorln("Can't get cron jobs, will not send SubScanJobStatus")
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
			ch.log.Warningln("Wait over 5 min, CronHelper is timeout")
		case <-nowContext.Done():
			ch.log.Infoln("CronHelper.Stop() Done.")
		}
	} else {
		// Stop stops the cron scheduler if it is running; otherwise it does nothing.
		// A context is returned so the caller can wait for running jobs to complete.
		nowContext := ch.c.Stop()
		select {
		case <-time.After(5 * time.Second):
			ch.log.Warningln("Wait over 5 s, CronHelper is timeout")
		case <-nowContext.Done():
			ch.log.Infoln("CronHelper.Stop() Done.")
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

		ch.log.Infoln(log_helper.OnceSubsScanEnd)

		// 下载完后，应该继续是等待
		tttt := ch.c.Entries()[0].Next.Format("2006-01-02 15:04:05")
		common.SetSubScanJobStatusWaiting(tttt)
	}()

	ch.fullSubDownloadProcessingLock.Lock()
	ch.fullSubDownloadProcessing = true
	ch.fullSubDownloadProcessingLock.Unlock()

	// ------------------------------------------------------------------------
	// 如果是 Debug 模式，那么就需要写入特殊文件
	if ch.sets.AdvancedSettings.DebugMode == true {
		err := log_helper.WriteDebugFile()
		if err != nil {
			ch.log.Errorln("log_helper.WriteDebugFile " + err.Error())
		}
		log_helper.GetLogger(true).Infoln("Reload Log Settings, level = Debug")
	} else {
		err := log_helper.DeleteDebugFile()
		if err != nil {
			ch.log.Errorln("log_helper.DeleteDebugFile " + err.Error())
		}
		log_helper.GetLogger(true).Infoln("Reload Log Settings, level = Info")
	}
	// ------------------------------------------------------------------------
	// 开始标记，这个是单次扫描的开始
	ch.log.Infoln(log_helper.OnceSubsScanStart)

	// 扫描字幕任务开始，先是扫描阶段，那么是拿不到有多少视频需要扫描的数量的
	common.SetSubScanJobStatusPreparing(time.Now().Format("2006-01-02 15:04:05"))

	// 下载前的初始化
	preDownloadProcess := pre_download_process.NewPreDownloadProcess(ch.log, ch.sets)
	err := preDownloadProcess.
		Init().
		Check().
		HotFix().
		ChangeSubNameFormat().
		ReloadBrowser().
		Wait()
	if err != nil {
		ch.log.Errorln("pre_download_process", "Error:", err)
		ch.log.Errorln("Skip DownloaderHelper.Start()")
		return
	}
	// 开始下载
	ch.dh = downloader_helper.NewDownloaderHelper(*settings.GetSettings(true),
		preDownloadProcess.SubSupplierHub)
	err = ch.dh.Start()
	if err != nil {
		ch.log.Errorln("downloader_helper.Start()", "Error:", err)
		return
	}
}

const (
	Stopped = "stopped"
	Running = "running"
)
