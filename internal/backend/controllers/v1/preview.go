package v1

import (
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/jinzhu/now"
	PTN "github.com/middelink/go-parse-torrent-name"
	"net/http"
	"strconv"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/preview_queue"
	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

// PreviewAdd 添加需要预览的任务,弃用
func (cb *ControllerBase) PreviewAdd(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewAdd", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	// 暂时不支持蓝光的预览
	if pkg.IsFile(job.VideoFPath) == false {
		bok, _, _ := decode.IsFakeBDMVWorked(job.VideoFPath)
		if bok == true {
			c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "not support blu-ray preview"})
			return
		} else {
			c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "video file not found"})
			return
		}
	}

	cb.cronHelper.Downloader.PreviewQueue.Add(&job)
	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}

// PreviewList 列举预览任务,弃用
func (cb *ControllerBase) PreviewList(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewList", err)
	}()

	listJob := cb.cronHelper.Downloader.PreviewQueue.ListJob()
	c.JSON(http.StatusOK, preview_queue.Reply{
		Jobs: listJob,
	})
}

// PreviewIsJobInQueue 预览的任务是否在列表中，或者说是在执行中,弃用
func (cb *ControllerBase) PreviewIsJobInQueue(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewIsJobInQueue", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	found := cb.cronHelper.Downloader.PreviewQueue.IsJobInQueue(&preview_queue.Job{
		VideoFPath: job.VideoFPath,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: strconv.FormatBool(found)})
	return
}

// PreviewJobResult 预览的任务的结果，成功 ok，不存在空，其他是失败,弃用
func (cb *ControllerBase) PreviewJobResult(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewJobResult", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	result := cb.cronHelper.Downloader.PreviewQueue.JobResult(&preview_queue.Job{
		VideoFPath: job.VideoFPath,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: result})
	return
}

// PreviewGetExportInfo 预览的任务的导出信息,弃用
func (cb *ControllerBase) PreviewGetExportInfo(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewGetExportInfo", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	m3u8, subPaths, err := cb.cronHelper.Downloader.PreviewQueue.GetVideoHLSAndSubByTimeRangeExportPathInfo(job.VideoFPath, job.SubFPaths, job.StartTime, job.EndTime)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, preview_queue.Job{
		VideoFPath: m3u8,
		SubFPaths:  subPaths,
	})
	return
}

func (cb *ControllerBase) PreviewCleanUp(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewCleanUp", err)
	}()

	if len(cb.cronHelper.Downloader.PreviewQueue.ListJob()) > 0 {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "false"})
		return
	}

	err = pkg.ClearVideoAndSubPreviewCacheFolder()
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "true"})
	return
}

func (cb *ControllerBase) PreviewSearchOtherWeb(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewSearchOtherWeb", err)
	}()

	searchOtherWeb := SearchOtherWebReq{}
	err = c.ShouldBindJSON(&searchOtherWeb)
	if err != nil {
		return
	}

	if pkg.IsFile(searchOtherWeb.VideoFPath) == false {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "video file not found"})
		return
	}

	mixMediaInfo, err := mix_media_info.GetMixMediaInfo(cb.cronHelper.FileDownloader.MediaInfoDealers,
		searchOtherWeb.VideoFPath, searchOtherWeb.IsMovie)
	if err != nil {
		return
	}
	// 搜索网站的地址
	searchOtherWebReply := SearchOtherWebReply{}
	searchOtherWebReply.SearchUrls = make([]string, 0)
	searchOtherWebReply.SearchUrls = append(searchOtherWebReply.SearchUrls, settings.Get().AdvancedSettings.SuppliersSettings.Zimuku.GetSearchUrl())
	searchOtherWebReply.SearchUrls = append(searchOtherWebReply.SearchUrls, settings.Get().AdvancedSettings.SuppliersSettings.SubHD.GetSearchUrl())
	searchOtherWebReply.SearchUrls = append(searchOtherWebReply.SearchUrls, settings.Get().AdvancedSettings.SuppliersSettings.A4k.GetSearchUrl())

	year, err := now.Parse(mixMediaInfo.Year)
	if err != nil {
		return
	}
	strYear := fmt.Sprintf("%d", year.Year())
	// 返回多种关键词
	searchOtherWebReply.KeyWords = make([]string, 0)
	// imdb id
	searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.ImdbId)

	if searchOtherWeb.IsMovie == true {
		// 电影
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn)
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn+" "+strYear)
		if mixMediaInfo.TitleCn != mixMediaInfo.TitleEn {
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn+" "+strYear)
		}
		if mixMediaInfo.TitleCn != mixMediaInfo.OriginalTitle && mixMediaInfo.OriginalTitle != mixMediaInfo.TitleEn {
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle+" "+strYear)
		}
	} else {
		// 电视剧
		var ptn *PTN.TorrentInfo
		ptn, err = decode.GetVideoInfoFromFileName(searchOtherWeb.VideoFPath)
		if err != nil {
			return
		}
		seasonKeyWord0 := " 第" + zh.Uint64(ptn.Season).String() + "季"
		seasonKeyWord1 := fmt.Sprintf(" S%02d", ptn.Season)
		seasonKeyWord2 := " " + pkg.GetEpisodeKeyName(ptn.Season, ptn.Episode, true)
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn)
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn+seasonKeyWord0)
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn+seasonKeyWord1)
		searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleCn+seasonKeyWord2)
		if mixMediaInfo.TitleCn != mixMediaInfo.TitleEn {
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn+seasonKeyWord0)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn+seasonKeyWord1)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.TitleEn+seasonKeyWord2)
		}
		if mixMediaInfo.TitleCn != mixMediaInfo.OriginalTitle && mixMediaInfo.OriginalTitle != mixMediaInfo.TitleEn {
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle+seasonKeyWord0)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle+seasonKeyWord1)
			searchOtherWebReply.KeyWords = append(searchOtherWebReply.KeyWords, mixMediaInfo.OriginalTitle+seasonKeyWord2)
		}
	}

	c.JSON(http.StatusOK, searchOtherWebReply)
}

func (cb *ControllerBase) PreviewVideoFPath2IMDBInfo(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewVideoFPath2IMDBInfo", err)
	}()

	searchOtherWeb := SearchOtherWebReq{}
	err = c.ShouldBindJSON(&searchOtherWeb)
	if err != nil {
		return
	}

	if pkg.IsFile(searchOtherWeb.VideoFPath) == false {
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "video file not found"})
		return
	}

	mixMediaInfo, err := mix_media_info.GetMixMediaInfo(cb.cronHelper.FileDownloader.MediaInfoDealers,
		searchOtherWeb.VideoFPath, searchOtherWeb.IsMovie)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, &mixMediaInfo)
}

type SearchOtherWebReq struct {
	VideoFPath string `json:"video_f_path"`
	IsMovie    bool   `json:"is_movie"`
}

type SearchOtherWebReply struct {
	KeyWords   []string `json:"key_words"`
	SearchUrls []string `json:"search_url"`
}
