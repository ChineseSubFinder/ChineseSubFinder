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
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Downloader struct {
	reqParam      common.ReqParam
	log           *logrus.Logger
	topic         int                        // 最多能够下载 Top 几的字幕，每一个网站
	mk            *mark_system.MarkingSystem // MarkingSystem
}

func NewDownloader(_reqParam ...common.ReqParam) *Downloader {

	var downloader Downloader
	downloader.log = model.GetLogger()
	downloader.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		downloader.reqParam = _reqParam[0]
		if downloader.reqParam.Topic > 0 && downloader.reqParam.Topic != downloader.topic {
			downloader.topic = downloader.reqParam.Topic
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
		// 抉择完毕，需要清理缓存目录
		err := model.ClearTmpFolder()
		if err != nil {
			d.log.Error(err)
		}
	}()
	nowVideoList, err := model.SearchMatchedVideoFile(dir)
	if err != nil {
		return err
	}
	// 构建每个字幕站点下载者的实例
	var subSupplierHub *sub_supplier.SubSupplierHub
	subSupplierHub = sub_supplier.NewSubSupplierHub(shooter.NewSupplier(d.reqParam),
		subhd.NewSupplier(d.reqParam),
		xunlei.NewSupplier(d.reqParam),
		zimuku.NewSupplier(d.reqParam),
	)
	// TODO 后续再改为每个视频以上的流程都是一个 channel 来做（目前做不了，得重构缓存字幕的方式，不然会出问题），并且需要控制在一个并发量之下（很可能没必要，毕竟要在弱鸡机器上挂机用的）
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range nowVideoList {
		// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
		organizeSubFiles, err := subSupplierHub.DownloadSub4Movie(oneVideoFullPath, i)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Movie", oneVideoFullPath ,err)
			continue
		}
		d.oneVideoSelectBestSub(oneVideoFullPath, organizeSubFiles)
		// -----------------------------------------------------
	}
	return nil
}

func (d Downloader) DownloadSub4Series(dir string) error {
	defer func() {
		// 抉择完毕，需要清理缓存目录
		err := model.ClearTmpFolder()
		if err != nil {
			d.log.Error(err)
		}
	}()
	// 构建每个字幕站点下载者的实例
	var subSupplierHub *sub_supplier.SubSupplierHub
	subSupplierHub = sub_supplier.NewSubSupplierHub(zimuku.NewSupplier(d.reqParam),
		//shooter.NewSupplier(d.reqParam),
		//subhd.NewSupplier(d.reqParam),
		//xunlei.NewSupplier(d.reqParam),
	)
	// 遍历连续剧总目录下的第一层目录
	seriesDirList, err := series_helper.GetSeriesList(dir)
	if err != nil {
		return err
	}
	for i, oneSeriesPath := range seriesDirList {
		// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
		seriesInfo, organizeSubFiles, err := subSupplierHub.DownloadSub4Series(oneSeriesPath, i)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Series", oneSeriesPath ,err)
			return err
		}

		// 只针对需要下载字幕的视频进行字幕的选择保存
		for epsKey, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
			// 匹配对应的 Eps 去处理
			d.oneVideoSelectBestSub(episodeInfo.FileFullPath, organizeSubFiles[epsKey])
		}
		// 这里会拿到一份季度字幕的列表比如，S1E0 S2E0 S3E0
		d.saveFullSeasonSub(seriesInfo, organizeSubFiles)
	}
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
func (d Downloader) saveFullSeasonSub(seriesInfo *common.SeriesInfo, organizeSubFiles map[string][]string) {

	for _, season := range seriesInfo.SeasonDict {
		epsKey := model.GetEpisodeKeyName(season, 0)
		subs, ok := organizeSubFiles[epsKey]
		if ok == false {
			continue
		}
		for _, sub := range subs {
			subFileName := filepath.Base(sub)
			newSeasonSubRootPath := path.Join(seriesInfo.DirPath, epsKey)
			_ = os.MkdirAll(newSeasonSubRootPath, os.ModePerm)
			newSubFullPath := path.Join(newSeasonSubRootPath, subFileName)
			err := os.Rename(sub, newSubFullPath)
			if err != nil {
				d.log.Errorln("saveFullSeasonSub", subFileName, err)
				continue
			}
		}
	}
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

