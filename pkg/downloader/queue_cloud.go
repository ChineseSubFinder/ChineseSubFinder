package downloader

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

func (d *Downloader) queueDownloaderCloud() {

	if pkg.LiteMode() == true ||
		settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false ||
		d.needSkipCloudTask == true {
		// Lite 版本无法执行复杂的任务，因为没有浏览器
		// 如果没有开启共享字幕也不会触发获取云端任务的逻辑
		return
	}

	// 查询云端是否有任务
	//nowInfo := dao.UpdateInfo(global_value.AppVersion(), settings.Get())
	//askDownloadTaskReply, err := d.fileDownloader.SubtitleBestApi.AskDownloadTask(nowInfo.Id)
	//if err != nil {
	//	d.log.Errorf("queueDownloaderCloud AskDownloadTask error: %v", err)
	//	return
	//}
	//if askDownloadTaskReply.Status == 0 {
	//	// 失败
	//	if askDownloadTaskReply.Message == "app version is too low" {
	//		// 版本过低，不能获取任务
	//		d.needSkipCloudTask = true
	//		d.log.Warnf("queueDownloaderCloud AskDownloadTask error: %v", askDownloadTaskReply.Message)
	//		return
	//	}
	//} else if askDownloadTaskReply.Status == 2 {
	//	// 没有任务
	//	return
	//}
	// 如果收到任务，那么就启动下载，这里下载的任务需要单独地方临时存储

	// 如果下载成功就反馈到云端
}
