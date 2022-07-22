package downloader

import "github.com/allanpk716/ChineseSubFinder/pkg/global_value"

func (d *Downloader) queueDownloaderCloud() {

	if global_value.LiteMode() == true || d.settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		// Lite 版本无法执行复杂的任务，因为没有浏览器
		// 如果没有开启共享字幕也不会出发获取云端任务的逻辑
		return
	}

	// 查询云端是否又任务
	//d.fileDownloader.SubtitleBestApi.

	// 如果收到任务，那么就启动下载，这里下载的任务需要单独地方临时存储

	// 如果下载成功就反馈到云端
}
