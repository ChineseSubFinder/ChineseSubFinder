package downloader

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/preview_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/manual_upload_sub_2_local"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/save_sub_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/scan_logic"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/assrt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	embyHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/emby_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	markSystem "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/mark_system"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/pre_download_process"
	subSupplier "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_timeline_fixer"
	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	subCommon "github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_queue"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Downloader 实例化一次用一次，不要反复的使用，很多临时标志位需要清理。
type Downloader struct {
	log                      *logrus.Logger
	fileDownloader           *file_downloader.FileDownloader
	ctx                      context.Context
	cancel                   context.CancelFunc
	subSupplierHub           *subSupplier.SubSupplierHub                      // 字幕提供源的集合，这个需要定时进行扫描，这些字幕源是否有效，以及下载验证码信息
	mk                       *markSystem.MarkingSystem                        // MarkingSystem，字幕的评价系统
	subFormatter             ifaces.ISubFormatter                             // 字幕格式化命名的实现
	subNameFormatter         subCommon.FormatterName                          // 从 inSubFormatter 推断出来
	subTimelineFixerHelperEx *sub_timeline_fixer.SubTimelineFixerHelperEx     // 字幕时间轴校正
	downloaderLock           sync.Mutex                                       // 取消执行 task control 的 Lock
	downloadQueue            *task_queue.TaskQueue                            // 需要下载的视频的队列
	embyHelper               *embyHelper.EmbyHelper                           // Emby 的实例
	ScanLogic                *scan_logic.ScanLogic                            // 是否扫描逻辑
	SaveSubHelper            *save_sub_helper.SaveSubHelper                   // 保存字幕的逻辑
	ManualUploadSub2Local    *manual_upload_sub_2_local.ManualUploadSub2Local // 手动上传字幕到本地
	PreviewQueue             *preview_queue.PreviewQueue                      // 预览队列

	cacheLocker   sync.Mutex
	movieInfoMap  map[string]MovieInfo  // 给 Web 界面使用的，Key: VideoFPath
	seasonInfoMap map[string]SeasonInfo // 给 Web 界面使用的,Key: RootDirPath

	needSkipCloudTask bool // 是否跳过云端任务，比如当前的 App 版本低于服务器的要求（过低可能爬虫已经失效，意义不大）
}

func NewDownloader(inSubFormatter ifaces.ISubFormatter, fileDownloader *file_downloader.FileDownloader, downloadQueue *task_queue.TaskQueue) *Downloader {

	var downloader Downloader
	downloader.fileDownloader = fileDownloader
	downloader.subFormatter = inSubFormatter
	downloader.fileDownloader = fileDownloader
	downloader.log = fileDownloader.Log
	// 参入设置信息
	// 检测是否某些参数超出范围
	settings.Get().Check()
	// 这里就不单独弄一个 settings.SubNameFormatter 字段来传递值了，因为 inSubFormatter 就已经知道是什么 formatter 了
	downloader.subNameFormatter = subCommon.FormatterName(downloader.subFormatter.GetFormatterFormatterName())

	var sitesSequence = make([]string, 0)
	// TODO 这里写固定了抉择字幕的顺序
	sitesSequence = append(sitesSequence, common2.SubSiteSubtitleBest)
	sitesSequence = append(sitesSequence, common2.SubSiteAssrt)
	sitesSequence = append(sitesSequence, common2.SubSiteA4K)
	sitesSequence = append(sitesSequence, common2.SubSiteShooter)
	sitesSequence = append(sitesSequence, common2.SubSiteXunLei)
	downloader.mk = markSystem.NewMarkingSystem(downloader.log, sitesSequence, settings.Get().AdvancedSettings.SubTypePriority)

	// 初始化，字幕校正的实例
	downloader.subTimelineFixerHelperEx = sub_timeline_fixer.NewSubTimelineFixerHelperEx(downloader.log, *settings.Get().TimelineFixerSettings)

	if settings.Get().AdvancedSettings.FixTimeLine == true {
		downloader.subTimelineFixerHelperEx.Check()
	}
	// 任务队列
	downloader.downloadQueue = downloadQueue
	// 单个任务的超时设置
	downloader.ctx, downloader.cancel = context.WithCancel(context.Background())
	// 用于字幕下载后的刷新
	if settings.Get().EmbySettings.Enable == true {
		downloader.embyHelper = embyHelper.NewEmbyHelper(downloader.fileDownloader.MediaInfoDealers)
	}

	downloader.ScanLogic = scan_logic.NewScanLogic(downloader.log)

	downloader.SaveSubHelper = save_sub_helper.NewSaveSubHelper(
		downloader.log,
		downloader.subFormatter,
		downloader.subTimelineFixerHelperEx)

	downloader.ManualUploadSub2Local = manual_upload_sub_2_local.NewManualUploadSub2Local(downloader.log, downloader.SaveSubHelper, downloader.ScanLogic)
	downloader.PreviewQueue = preview_queue.NewPreviewQueue(downloader.log)

	downloader.movieInfoMap = make(map[string]MovieInfo)
	downloader.seasonInfoMap = make(map[string]SeasonInfo)

	err := downloader.loadVideoListCache()
	if err != nil {
		downloader.log.Errorln("loadVideoListCache error:", err)
	}

	return &downloader
}

// SupplierCheck 检查字幕源是否有效，会影响后续的字幕源是否参与下载
func (d *Downloader) SupplierCheck() {

	defer func() {
		if p := recover(); p != nil {
			d.log.Errorln("Downloader.SupplierCheck() panic")
			pkg.PrintPanicStack(d.log)
		}
		d.downloaderLock.Unlock()

		d.log.Infoln("Download.SupplierCheck() End")
	}()

	d.downloaderLock.Lock()
	d.log.Infoln("Download.SupplierCheck() Start ...")

	//// 创建一个 chan 用于任务的中断和超时
	//done := make(chan interface{}, 1)
	//// 接收内部任务的 panic
	//panicChan := make(chan interface{}, 1)
	//
	//go func() {
	//	defer func() {
	//		if p := recover(); p != nil {
	//			panicChan <- p
	//		}
	//
	//		close(done)
	//		close(panicChan)
	//	}()
	// 下载前的初始化
	d.log.Infoln("PreDownloadProcess.Init().Check().Wait()...")

	if settings.Get().SpeedDevMode == true {
		// 这里是调试使用的，指定了只用一个字幕源
		//subSupplierHub := subSupplier.NewSubSupplierHub(csf.NewSupplier(d.fileDownloader))
		subSupplierHub := subSupplier.NewSubSupplierHub(assrt.NewSupplier(d.fileDownloader))
		d.subSupplierHub = subSupplierHub
	} else {

		preDownloadProcess := pre_download_process.NewPreDownloadProcess(d.fileDownloader)
		err := preDownloadProcess.Init().Check().Wait()
		if err != nil {
			//done <- errors.New(fmt.Sprintf("NewPreDownloadProcess Error: %v", err))
			d.log.Errorln(errors.New(fmt.Sprintf("NewPreDownloadProcess Error: %v", err)))
		} else {
			// 更新 SubSupplierHub 实例
			d.subSupplierHub = preDownloadProcess.SubSupplierHub
			//done <- nil
		}
	}

	//	done <- nil
	//}()
	//
	//select {
	//case err := <-done:
	//	if err != nil {
	//		d.log.Errorln(err)
	//	}
	//	break
	//case p := <-panicChan:
	//	// 遇到内部的 panic，向外抛出
	//	panic(p)
	//case <-d.ctx.Done():
	//	{
	//		d.log.Errorln("cancel SupplierCheck")
	//		return
	//	}
	//}
}

// QueueDownloader 从字幕队列中取一个视频的字幕下载任务出来，并且开始下载
func (d *Downloader) QueueDownloader() {

	// 本地的任务
	d.queueDownloaderLocal()
	// 云端分布式的任务
	d.queueDownloaderCloud()
}

func (d *Downloader) Cancel() {
	if d == nil {
		return
	}
	d.cancel()
	d.log.Infoln("Downloader.Cancel()")
}
