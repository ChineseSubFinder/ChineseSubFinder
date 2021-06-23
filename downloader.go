package main

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/mark_system"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/series_helper"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/zimuku"
	"github.com/go-rod/rod/lib/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Downloader struct {
	reqParam      common.ReqParam
	log           *logrus.Logger
	topic         int                        // 最多能够下载 Top 几的字幕，每一个网站
	mk            *mark_system.MarkingSystem // MarkingSystem
}

// TODO 加入一个功能，如果这个电影，发行很久了，而且没得字幕，那么重复多天后，应该就不用再扫描它了 把字幕下载周期开放出来允许设置，现在默认是 3个月

func NewDownloader(_reqParam ...common.ReqParam) *Downloader {

	var downloader Downloader
	downloader.log = model.GetLogger()
	downloader.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		downloader.reqParam = _reqParam[0]
		if downloader.reqParam.Topic > 0 && downloader.reqParam.Topic != downloader.topic {
			downloader.topic = downloader.reqParam.Topic
		}
		// 并发线程的范围控制
		if downloader.reqParam.Threads <= 0 {
			downloader.reqParam.Threads = 2
		} else if downloader.reqParam.Threads >= 10 {
			downloader.reqParam.Threads = 10
		}
	}

	var sitesSequence = make([]string, 0)
	// TODO 这里写固定了抉择字幕的顺序
	sitesSequence = append(sitesSequence, common.SubSiteZiMuKu)
	sitesSequence = append(sitesSequence, common.SubSiteSubHd)
	sitesSequence = append(sitesSequence, common.SubSiteXunLei)
	sitesSequence = append(sitesSequence, common.SubSiteShooter)
	downloader.mk = mark_system.NewMarkingSystem(sitesSequence)

	return &downloader
}

func (d Downloader) DownloadSub4Movie(dir string) error {
	defer func() {
		// 所有的电影字幕下载完成，抉择完成，需要清理缓存目录
		err := model.ClearRootTmpFolder()
		if err != nil {
			d.log.Error(err)
		}
		// 注意并发 pool 的释放
		defer ants.Release()
	}()

	nowVideoList, err := model.SearchMatchedVideoFile(dir)
	if err != nil {
		return err
	}

	// 并发控制
	movieDlFunc := func(i interface{}) {
		inData := i.(InputData)
		// -----------------------------------------------------
		// 构建每个字幕站点下载者的实例
		var subSupplierHub = sub_supplier.NewSubSupplierHub(
			shooter.NewSupplier(d.reqParam),
			subhd.NewSupplier(d.reqParam),
			xunlei.NewSupplier(d.reqParam),
			zimuku.NewSupplier(d.reqParam),
		)
		// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
		organizeSubFiles, err := subSupplierHub.DownloadSub4Movie(inData.OneVideoFullPath, inData.Index)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Movie", inData.OneVideoFullPath ,err)
			return
		}
		if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
			d.log.Infoln("no sub found", filepath.Base(inData.OneVideoFullPath))
			return
		}
		d.oneVideoSelectBestSub(inData.OneVideoFullPath, organizeSubFiles)
		// -----------------------------------------------------
	}
	wg := sync.WaitGroup{}
	p, err := ants.NewPoolWithFunc(d.reqParam.Threads, func(inData interface{}) {
		movieDlFunc(inData)
		wg.Done()
	})
	if err != nil {
		return err
	}
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range nowVideoList {
		wg.Add(1)
		err = p.Invoke(InputData{OneVideoFullPath: oneVideoFullPath, Index: i})
		if err != nil {
			d.log.Errorln("movie ants.Invoke",err)
		}
	}
	wg.Wait()
	return nil
}

func (d Downloader) DownloadSub4Series(dir string) error {
	defer func() {
		// 所有的连续剧字幕下载完成，抉择完成，需要清理缓存目录
		err := model.ClearRootTmpFolder()
		if err != nil {
			d.log.Error(err)
		}
		// 注意并发 pool 的释放
		defer ants.Release()
	}()

	// 并发控制
	seriesDlFunc := func(i interface{}) {

		inData := i.(InputData)
		// 构建每个字幕站点下载者的实例
		var subSupplierHub *sub_supplier.SubSupplierHub
		subSupplierHub = sub_supplier.NewSubSupplierHub(
			zimuku.NewSupplier(d.reqParam),
			subhd.NewSupplier(d.reqParam),
			xunlei.NewSupplier(d.reqParam),
			shooter.NewSupplier(d.reqParam),
		)
		// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
		seriesInfo, organizeSubFiles, err := subSupplierHub.DownloadSub4Series(inData.OneVideoFullPath, inData.Index)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Series", inData.OneVideoFullPath ,err)
			return
		}
		if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
			d.log.Infoln("no sub found", filepath.Base(inData.OneVideoFullPath))
			return
		}
		// 只针对需要下载字幕的视频进行字幕的选择保存
		for epsKey, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
			// 匹配对应的 Eps 去处理
			d.oneVideoSelectBestSub(episodeInfo.FileFullPath, organizeSubFiles[epsKey])
		}
		// 这里会拿到一份季度字幕的列表比如，Key 是 S1E0 S2E0 S3E0，value 是新的存储位置
		fullSeasonSubDict := d.saveFullSeasonSub(seriesInfo, organizeSubFiles)
		// TODO 季度的字幕包，应该优先于零散的字幕吧，暂定就这样了，注意是全部都替换
		// 需要与有下载需求的季交叉
		for _, episodeInfo := range seriesInfo.EpList {
			_, ok := seriesInfo.NeedDlSeasonDict[episodeInfo.Season]
			if ok == false {
				continue
			}
			// 匹配对应的 Eps 去处理
			seasonEpsKey := model.GetEpisodeKeyName(episodeInfo.Season, episodeInfo.Episode)
			d.oneVideoSelectBestSub(episodeInfo.FileFullPath, fullSeasonSubDict[seasonEpsKey])
		}
	}
	wg := sync.WaitGroup{}
	p, err := ants.NewPoolWithFunc(d.reqParam.Threads, func(inData interface{}) {
		seriesDlFunc(inData)
		wg.Done()
	})
	if err != nil {
		return err
	}
	// 遍历连续剧总目录下的第一层目录
	seriesDirList, err := series_helper.GetSeriesList(dir)
	if err != nil {
		return err
	}
	for i, oneSeriesPath := range seriesDirList {
		wg.Add(1)
		err = p.Invoke(InputData{OneVideoFullPath: oneSeriesPath, Index: i})
		if err != nil {
			d.log.Errorln("series ants.Invoke",err)
		}
	}
	wg.Wait()
	return nil
}

// oneVideoSelectBestSub 一个视频，选择最佳的一个字幕（也可以保存所有网站第一个最佳字幕）
func (d Downloader) oneVideoSelectBestSub(oneVideoFullPath string, organizeSubFiles []string) {
	var err error
	// 得到目标视频文件的根目录
	videoRootPath := filepath.Dir(oneVideoFullPath)
	// -------------------------------------------------
	// 调试缓存，把下载好的字幕写到对应的视频目录下，方便调试
	if d.reqParam.DebugMode == true {
		err = d.copySubFile2DesFolder(videoRootPath, organizeSubFiles)
		if err != nil {
			d.log.Errorln("copySubFile2DesFolder", err)
		}
	}
	// -------------------------------------------------
	if d.reqParam.SaveMultiSub == false {
		// 选择最优的一个字幕
		var finalSubFile *common.SubParserFileInfo
		finalSubFile = d.mk.SelectOneSubFile(organizeSubFiles)
		if finalSubFile == nil {
			d.log.Warnln("Found", len(organizeSubFiles), " subtitles but not one fit:", oneVideoFullPath)
			return
		}
		// 找到了，写入文件
		err = d.writeSubFile2VideoPath(oneVideoFullPath, *finalSubFile, "")
		if err != nil {
			d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub, "writeSubFile2VideoPath:", err)
			return
		}
	} else {
		// 每个网站 Top1 的字幕
		siteNames, finalSubFiles := d.mk.SelectEachSiteTop1SubFile(organizeSubFiles)
		if len(siteNames) < 0 {
			d.log.Warnln("SelectEachSiteTop1SubFile found none sub file")
			return
		}

		for i, file := range finalSubFiles {
			err = d.writeSubFile2VideoPath(oneVideoFullPath, file, siteNames[i])
			if err != nil {
				d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub, "writeSubFile2VideoPath:", err)
				return
			}
		}
	}
}

// saveFullSeasonSub 这里就需要单独存储到连续剧每一季的文件夹的特殊文件夹中
func (d Downloader) saveFullSeasonSub(seriesInfo *common.SeriesInfo, organizeSubFiles map[string][]string) map[string][]string {

	var fullSeasonSubDict = make(map[string][]string)

	for _, season := range seriesInfo.SeasonDict {
		seasonKey := model.GetEpisodeKeyName(season, 0)
		subs, ok := organizeSubFiles[seasonKey]
		if ok == false {
			continue
		}
		for _, sub := range subs {
			subFileName := filepath.Base(sub)
			newSeasonSubRootPath := path.Join(seriesInfo.DirPath, "Sub_"+seasonKey)
			_ = os.MkdirAll(newSeasonSubRootPath, os.ModePerm)
			newSubFullPath := path.Join(newSeasonSubRootPath, subFileName)
			_, err := model.CopyFile(newSubFullPath, sub)
			if err != nil {
				d.log.Errorln("saveFullSeasonSub", subFileName, err)
				continue
			}
			// 从字幕的文件名推断是 哪一季 的 那一集
			_, gusSeason, gusEpisode, err := model.GetSeasonAndEpisodeFromSubFileName(subFileName)
			if err != nil {
				return nil
			}
			// 把整季的字幕缓存位置也提供出去，如果之前没有下载到的，这里返回出来的可以补上
			seasonEpsKey := model.GetEpisodeKeyName(gusSeason, gusEpisode)
			_, ok := fullSeasonSubDict[seasonEpsKey]
			if ok == false {
				// 初始化
				fullSeasonSubDict[seasonEpsKey] = make([]string, 0)
			}
			fullSeasonSubDict[seasonEpsKey] = append(fullSeasonSubDict[seasonEpsKey], sub)
		}
	}

	return fullSeasonSubDict
}

// 在前面需要进行语言的筛选、排序，这里仅仅是存储
func (d Downloader) writeSubFile2VideoPath(videoFileFullPath string, finalSubFile common.SubParserFileInfo, extraSubPreName string) error {
	videoRootPath := filepath.Dir(videoFileFullPath)
	embyLanExtName := model.Lang2EmbyName(finalSubFile.Lang)
	// 构建视频文件加 emby 的字幕预研要求名称
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileFullPath),
		filepath.Ext(videoFileFullPath), "")
	if extraSubPreName != "" {
		extraSubPreName = "[" + extraSubPreName +"]"
	}
	subNewName := videoFileNameWithOutExt + embyLanExtName + extraSubPreName + finalSubFile.Ext
	desSubFullPath := path.Join(videoRootPath, subNewName)
	// 最后写入字幕
	err := utils.OutputFile(desSubFullPath, finalSubFile.Data)
	if err != nil {
		return err
	}
	d.log.Infoln("OrgSubName:", finalSubFile.Name)
	d.log.Infoln("SubDownAt:", desSubFullPath)
	return nil
}

// copySubFile2DesFolder 拷贝字幕文件到目标文件夹
func (d Downloader) copySubFile2DesFolder(desFolder string, subFiles []string) error {

	// 需要进行字幕文件的缓存
	// 把缓存的文件夹新建出来
	desFolderFullPath := path.Join(desFolder, common.SubTmpFolderName)
	err := os.MkdirAll(desFolderFullPath, os.ModePerm)
	if err != nil {
		return err
	}
	// 复制下载在 tmp 文件夹中的字幕文件到视频文件夹下面
	for _, subFile := range subFiles {
		newFn := path.Join(desFolderFullPath, filepath.Base(subFile))
		_, err = model.CopyFile(newFn, subFile)
		if err != nil {
			return err
		}
	}

	return nil
}

type InputData struct {
	OneVideoFullPath string
	Index			int
}