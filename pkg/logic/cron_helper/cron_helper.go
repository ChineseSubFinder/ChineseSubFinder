package cron_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	//"github.com/ChineseSubFinder/ChineseSubFinder/internal/logic/pre_job"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_queue"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/video_scan_and_refresh_helper"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type CronHelper struct {
	stopping          bool // 正在停止
	cronHelperRunning bool // 这个是定时器启动的状态，它为true，不代表核心函数在执行
	//scanPlayedVideoSubInfo    *scan_played_video_subinfo.ScanPlayedVideoSubInfo        // 扫描已经播放过的视频的字幕信息
	FileDownloader            *file_downloader.FileDownloader                          // 文件下载器
	DownloadQueue             *task_queue.TaskQueue                                    // 需要下载的视频的队列
	Downloader                *downloader.Downloader                                   // 下载者线程
	videoScanAndRefreshHelper *video_scan_and_refresh_helper.VideoScanAndRefreshHelper // 视频扫描和刷新的帮助类
	cronLock                  sync.Mutex                                               // 锁
	c                         *cron.Cron                                               // 定时器实例
	Logger                    *logrus.Logger                                           // 日志实例
	entryIDScanVideoProcess   cron.EntryID                                             // 建立视频缓存，扫描有那些视频需要进行字幕下载的定时器的 ID
	entryIDSupplierCheck      cron.EntryID                                             // 检查字幕源有效性的定时器的 ID
	entryIDQueueDownloader    cron.EntryID                                             // 下载队列的定时器的 ID
	entryIDFeedBack           cron.EntryID                                             // 信息反馈
	//entryIDScanPlayedVideoSubInfo cron.EntryID
	//entryIDUploadPlayedVideoSub cron.EntryID
}

func NewCronHelper(fileDownloader *file_downloader.FileDownloader) *CronHelper {

	ch := CronHelper{
		FileDownloader: fileDownloader,
		Logger:         fileDownloader.Log,
		// 实例化下载队列
		DownloadQueue: task_queue.NewTaskQueue(fileDownloader.CacheCenter),
	}

	//var err error
	// ----------------------------------------------
	// 扫描已播放
	//ch.scanPlayedVideoSubInfo, err = scan_played_video_subinfo.NewScanPlayedVideoSubInfo(ch.Logger, fileDownloader)
	//if err != nil {
	//	ch.Logger.Panicln(err)
	//}
	// ----------------------------------------------
	// 字幕扫描器
	ch.videoScanAndRefreshHelper = video_scan_and_refresh_helper.NewVideoScanAndRefreshHelper(
		sub_formatter.GetSubFormatter(ch.Logger, settings.Get().AdvancedSettings.SubNameFormatter),
		ch.FileDownloader,
		ch.DownloadQueue)

	// ----------------------------------------------
	// 初始化下载者，里面的两个 func 需要使用定时器启动 SupplierCheck QueueDownloader
	ch.Downloader = downloader.NewDownloader(
		sub_formatter.GetSubFormatter(ch.Logger, settings.Get().AdvancedSettings.SubNameFormatter),
		ch.FileDownloader, ch.DownloadQueue)

	// 强制进行一次字幕源有效性检查
	ch.Downloader.SupplierCheck()

	return &ch
}

// Start 开启定时器任务，这个任务是非阻塞的，scanVideoProcessAdd2DownloadQueue 仅仅可能是这个函数执行耗时而已
// runImmediately == false 那么 ch.c.Start() 是不会阻塞的
func (ch *CronHelper) Start(runImmediately bool) {

	ch.cronLock.Lock()
	if ch.cronHelperRunning == true {
		ch.cronLock.Unlock()
		return
	}
	ch.cronLock.Unlock()

	ch.cronLock.Lock()
	ch.cronHelperRunning = true
	ch.stopping = false
	ch.cronLock.Unlock()
	// ----------------------------------------------
	// 判断扫描任务的时间间隔是否符合要求，不符合则重写默认值
	_, err := cron.ParseStandard(settings.Get().CommonSettings.ScanInterval)
	if err != nil {
		ch.Logger.Warningln("CommonSettings.ScanInterval format error, after v0.25.x , need reset this at WebUI")
		// 如果解析错误了，就需要重新赋值默认值过来，然后保存
		nowSettings := settings.Get()
		nowSettings.CommonSettings.ScanInterval = settings.NewCommonSettings().ScanInterval
		err = settings.SetFullNewSettings(nowSettings)
		if err != nil {
			ch.Logger.Panicln("CronHelper.SetFullNewSettings:", err)
			return
		}
	}
	// ----------------------------------------------
	ch.c = cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
	{
		// 测试部分定时器代码，提前运行
		if settings.Get().SpeedDevMode == true {

			//pj := pre_job.NewPreJob(settings.Get(), ch.Logger)
			//err = pj.HotFix().Wait()
			//if err != nil {
			//	ch.Logger.Errorln(err)
			//}

			//ch.scanVideoProcessAdd2DownloadQueue()
			//
			//ch.scanPlayedVideoSub()

			//ch.uploadVideoSub()
		}
	}

	// 定时器
	// 这个暂时无法被取消执行
	ch.entryIDScanVideoProcess, err = ch.c.AddFunc(settings.Get().CommonSettings.ScanInterval, ch.scanVideoProcessAdd2DownloadQueue)
	if err != nil {
		ch.Logger.Panicln("CronHelper scanVideoProcessAdd2DownloadQueue, scanVideoProcessAdd2DownloadQueue Cron entryID:", ch.entryIDScanVideoProcess, "Error:", err)
	}
	// 这个可以由 ch.Downloader.Cancel() 取消执行
	ch.entryIDSupplierCheck, err = ch.c.AddFunc("@every 1h", ch.Downloader.SupplierCheck)
	if err != nil {
		ch.Logger.Panicln("CronHelper SupplierCheck, SupplierCheck Cron entryID:", ch.entryIDSupplierCheck, "Error:", err)
	}
	// 这个可以由 ch.Downloader.Cancel() 取消执行
	ch.entryIDQueueDownloader, err = ch.c.AddFunc("@every 15s", ch.Downloader.QueueDownloader)
	if err != nil {
		ch.Logger.Panicln("CronHelper QueueDownloader, QueueDownloader Cron entryID:", ch.entryIDQueueDownloader, "Error:", err)
	}
	// FeedBack
	ch.entryIDFeedBack, err = ch.c.AddFunc("@every 24h", ch.feedBack)
	if err != nil {
		ch.Logger.Panicln("CronHelper QueueDownloader, feedBack Cron entryID:", ch.entryIDFeedBack, "Error:", err)
	}
	// 字幕的上传逻辑
	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == true {
		// 取消上传字幕的逻辑，目前评估的第一阶段已经完成，后续的逻辑需要重新评估
		//intervalNowTask := "@every 10s"
		//if settings.Get().SpeedDevMode == true {
		//	intervalNowTask = "@every 1s"
		//}
		//ch.entryIDUploadPlayedVideoSub, err = ch.c.AddFunc(intervalNowTask, ch.uploadVideoSub)
		//if err != nil {
		//	ch.Logger.Panicln("CronHelper QueueDownloader, uploadVideoSub Cron entryID:", ch.entryIDUploadPlayedVideoSub, "Error:", err)
		//}
	}

	// ----------------------------------------------
	if runImmediately == true {
		// 是否在定时器开启前先执行一次视频扫描任务
		ch.cronLock.Lock()
		if ch.cronHelperRunning == true && ch.stopping == false {
			ch.cronLock.Unlock()
			//----------------------------------------------
			// 没有停止，那么继续扫描
			ch.Logger.Infoln("First Time scanVideoProcessAdd2DownloadQueue Start")
			if settings.Get().SpeedDevMode == false {
				ch.scanVideoProcessAdd2DownloadQueue()
			}
			ch.Logger.Infoln("First Time scanVideoProcessAdd2DownloadQueue End")
			//----------------------------------------------
		} else {
			ch.cronLock.Unlock()
			ch.Logger.Infoln("CronHelper is stopping, not start scanVideoProcessAdd2DownloadQueue")
			return
		}
	} else {
		ch.Logger.Infoln("RunAtStartup: false, so will not Run At Startup")
	}
	// ----------------------------------------------
	// 如果不是立即执行，那么就等待定时器开启
	ch.cronLock.Lock()
	if ch.cronHelperRunning == true && ch.stopping == false {
		ch.cronLock.Unlock()
		//----------------------------------------------
		ch.Logger.Infoln("CronHelper Start...")
		ch.c.Start()
		//----------------------------------------------
		// 只有定时任务 start 之后才能拿到信息
		if len(ch.c.Entries()) > 0 {
			// 不会马上启动扫描，那么就需要设置当前的时间，且为 waiting
			tttt := ch.c.Entry(ch.entryIDScanVideoProcess).Next.Format("2006-01-02 15:04:05")
			ch.Logger.Infoln("Next Sub Scan Will Process At:", tttt)
		} else {
			ch.Logger.Errorln("Can't get cron jobs, will not send SubScanJobStatus")
		}
		//----------------------------------------------
	} else {
		ch.cronLock.Unlock()
		ch.Logger.Infoln("CronHelper is stopping, not start CronHelper")
	}
	//----------------------------------------------

}

// Stop 会阻塞等待任务完成
func (ch *CronHelper) Stop() {

	cronHelperRunning := false
	ch.cronLock.Lock()
	cronHelperRunning = ch.cronHelperRunning
	ch.cronLock.Unlock()

	if cronHelperRunning == false {
		return
	}

	ch.cronLock.Lock()
	if ch.stopping == true {
		ch.cronLock.Unlock()
		return
	}
	ch.stopping = true
	ch.cronLock.Unlock()

	ch.videoScanAndRefreshHelper.Cancel()
	ch.Downloader.Cancel()
	//ch.scanPlayedVideoSubInfo.Cancel()
	// Stop stops the cron scheduler if it is running; otherwise it does nothing.
	// A context is returned so the caller can wait for running jobs to complete.
	nowContext := ch.c.Stop()
	select {
	case <-time.After(5 * time.Minute):
		ch.Logger.Warningln("Wait over 5 min, CronHelper is timeout")
	case <-nowContext.Done():
		ch.Logger.Infoln("CronHelper.Stop() context<-Done.")
	}

	ch.cronLock.Lock()
	ch.cronHelperRunning = false
	ch.stopping = false
	ch.cronLock.Unlock()

	ch.Logger.Infoln("CronHelper.Stop() Done.")
}

func (ch *CronHelper) feedBack() {
	ch.Logger.Infoln("Update Info...")
	nowInfo := dao.UpdateInfo(pkg.AppVersion(), settings.Get())
	_, err := ch.FileDownloader.MediaInfoDealers.SubtitleBestApi.FeedBack(
		nowInfo.Id,
		nowInfo.Version, nowInfo.MediaServer,
		nowInfo.EnableShare, nowInfo.EnableApiKey)
	if err != nil {
		ch.Logger.Errorln("FeedBack Error:", err)
		return
	}
}

//func (ch *CronHelper) scanPlayedVideoSub() {
//
//	ch.Logger.Infoln("Update Info...")
//	nowInfo := dao.UpdateInfo(pkg.AppVersion(), settings.Get())
//	_, err := ch.FileDownloader.MediaInfoDealers.SubtitleBestApi.FeedBack(
//		nowInfo.Id,
//		nowInfo.Version, nowInfo.MediaServer,
//		nowInfo.EnableShare, nowInfo.EnableApiKey)
//	if err != nil {
//		ch.Logger.Errorln("FeedBack Error:", err)
//		return
//	}
//
//	ch.Logger.Infoln("------------------------------------------------------")
//	ch.Logger.Infoln("scanPlayedVideoSub Start")
//	startT := time.Now()
//	defer func() {
//		ch.Logger.Infoln("scanPlayedVideoSub End, Cost:", time.Since(startT).Minutes(), "min")
//		ch.Logger.Infoln("------------------------------------------------------")
//	}()
//	bok, err := ch.scanPlayedVideoSubInfo.GetPlayedItemsSubtitle(settings.Get().EmbySettings, common.EmbyApiGetItemsLimitMax)
//	if err != nil {
//		ch.Logger.Errorln(err)
//	}
//	if bok == true {
//		ch.scanPlayedVideoSubInfo.Clear()
//		err = ch.scanPlayedVideoSubInfo.Scan()
//		if err != nil {
//			ch.Logger.Errorln(err)
//		}
//	}
//}

func (ch *CronHelper) CronHelperRunning() bool {

	defer func() {
		ch.cronLock.Unlock()
	}()
	ch.cronLock.Lock()
	return ch.cronHelperRunning
}

func (ch *CronHelper) CronHelperStopping() bool {

	defer func() {
		ch.cronLock.Unlock()
	}()
	ch.cronLock.Lock()
	return ch.stopping
}

func (ch *CronHelper) CronRunningStatusString() string {

	if ch.CronHelperRunning() == true {
		if ch.CronHelperStopping() == true {
			return Stopping
		}
		return Running
	} else {
		return Stopped
	}
}

// scanVideoProcessAdd2DownloadQueue 定时执行的视频扫描任务，提交给任务队列，然后由额外的下载者线程去取队列中的任务下载
func (ch *CronHelper) scanVideoProcessAdd2DownloadQueue() {

	// 扫描字幕任务开始，先是扫描阶段，那么是拿不到有多少视频需要扫描的数量的
	ch.Logger.Infoln("scanVideoProcessAdd2DownloadQueue Start:", time.Now().Format("2006-01-02 15:04:05"))
	// ----------------------------------------------------------------------------------------
	// ----------------------------------------------------------------------------------------
	// 扫描有那些视频需要下载字幕，放入队列中，然后会有下载者去这个队列取出来进行下载
	err := ch.videoScanAndRefreshHelper.Start(ch.Downloader.ScanLogic)
	if err != nil {
		ch.Logger.Errorln(err)
		return
	}
}

const (
	Stopped  = "stopped"
	Running  = "running"
	Stopping = "stopping"
)
