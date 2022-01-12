package internal

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/forced_scan_and_down_sub"
	markSystem "github.com/allanpk716/ChineseSubFinder/internal/logic/mark_system"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/restore_fix_timeline_bk"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	subcommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	sub_timeline_fixer_pkg "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os"
	"path/filepath"
	"sync"
)

// Downloader 实例化一次用一次，不要反复的使用，很多临时标志位需要清理。
type Downloader struct {
	reqParam                 types.ReqParam
	log                      *logrus.Logger
	topic                    int                       // 最多能够下载 Top 几的字幕，每一个网站
	mk                       *markSystem.MarkingSystem // MarkingSystem
	embyHelper               *embyHelper.EmbyHelper
	movieFileFullPathList    []string                      //  多个需要搜索字幕的电影文件全路径
	seriesSubNeedDlMap       map[string][]emby.EmbyMixInfo //  多个需要搜索字幕的连续剧目录
	subFormatter             ifaces.ISubFormatter          //	字幕格式化命名的实现
	subNameFormatter         subcommon.FormatterName       // 从 inSubFormatter 推断出来
	needForcedScanAndDownSub bool                          // 将会强制扫描所有的视频，下载字幕，替换已经存在的字幕，不进行时间段和已存在则跳过的判断。且不会进过 Emby API 的逻辑，智能进行强制去以本程序的方式去扫描。
	NeedRestoreFixTimeLineBK bool                          // 从 csf-bk 文件还原时间轴修复前的字幕文件

	subTimelineFixerHelperEx *sub_timeline_fixer.SubTimelineFixerHelperEx // 字幕时间轴校正
}

func NewDownloader(inSubFormatter ifaces.ISubFormatter, _reqParam ...types.ReqParam) *Downloader {

	var downloader Downloader
	downloader.subFormatter = inSubFormatter
	downloader.log = log_helper.GetLogger()
	downloader.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		downloader.reqParam = _reqParam[0]
		if downloader.reqParam.Topic > 0 && downloader.reqParam.Topic != downloader.topic {
			downloader.topic = downloader.reqParam.Topic
		}
		// 如果 Debug 模式开启了，强制设置线程数为1，方便定位问题
		if downloader.reqParam.DebugMode == true {
			downloader.reqParam.Threads = 1
		} else {
			// 并发线程的范围控制
			if downloader.reqParam.Threads <= 0 {
				downloader.reqParam.Threads = 2
			} else if downloader.reqParam.Threads >= 10 {
				downloader.reqParam.Threads = 10
			}
		}
		// 初始化 Emby API 接口
		if downloader.reqParam.EmbyConfig.Url != "" && downloader.reqParam.EmbyConfig.ApiKey != "" {
			downloader.embyHelper = embyHelper.NewEmbyHelper(downloader.reqParam.EmbyConfig)
		}
	} else {
		downloader.reqParam = *types.NewReqParam()
	}
	// 强制下载线程为 1，太猛，不然都是错误
	downloader.reqParam.Threads = 1

	// 这里就不单独弄一个 reqParam.SubNameFormatter 字段来传递值了，因为 inSubFormatter 就已经知道是什么 formatter 了
	downloader.subNameFormatter = subcommon.FormatterName(downloader.subFormatter.GetFormatterFormatterName())

	var sitesSequence = make([]string, 0)
	// TODO 这里写固定了抉择字幕的顺序
	sitesSequence = append(sitesSequence, common.SubSiteZiMuKu)
	sitesSequence = append(sitesSequence, common.SubSiteSubHd)
	sitesSequence = append(sitesSequence, common.SubSiteShooter)
	sitesSequence = append(sitesSequence, common.SubSiteXunLei)
	downloader.mk = markSystem.NewMarkingSystem(sitesSequence, downloader.reqParam.SubTypePriority)

	downloader.movieFileFullPathList = make([]string, 0)
	downloader.seriesSubNeedDlMap = make(map[string][]emby.EmbyMixInfo)

	// 初始化，字幕校正的实例
	downloader.subTimelineFixerHelperEx = sub_timeline_fixer.NewSubTimelineFixerHelperEx(downloader.reqParam.SubTimelineFixerConfig)

	if downloader.reqParam.FixTimeLine == true {
		downloader.subTimelineFixerHelperEx.Check()
	}

	return &downloader
}

// ReadSpeFile 优先级最高。读取特殊文件，启用一些特殊的功能，比如 forced_scan_and_down_sub
func (d *Downloader) ReadSpeFile() error {
	// 理论上是一次性的，用了这个文件就应该没了
	// 强制的字幕扫描
	needProcess_forced_scan_and_down_sub, err := forced_scan_and_down_sub.CheckSpeFile()
	if err != nil {
		return err
	}
	d.needForcedScanAndDownSub = needProcess_forced_scan_and_down_sub
	// 从 csf-bk 文件还原时间轴修复前的字幕文件
	needProcess_restore_fix_timeline_bk, err := restore_fix_timeline_bk.CheckSpeFile()
	if err != nil {
		return err
	}
	d.NeedRestoreFixTimeLineBK = needProcess_restore_fix_timeline_bk

	return nil
}

// GetUpdateVideoListFromEmby 这里首先会进行近期影片的获取，然后对这些影片进行刷新，然后在获取字幕列表，最终得到需要字幕获取的 video 列表
func (d *Downloader) GetUpdateVideoListFromEmby(movieRootDir, seriesRootDir string) error {
	if d.embyHelper == nil {
		return nil
	}
	var err error
	var movieList []emby.EmbyMixInfo
	movieList, d.seriesSubNeedDlMap, err = d.embyHelper.GetRecentlyAddVideoList(movieRootDir, seriesRootDir)
	if err != nil {
		return err
	}
	// 获取全路径
	for _, info := range movieList {
		d.movieFileFullPathList = append(d.movieFileFullPathList, info.VideoFileFullPath)
	}
	// 输出调试信息
	log_helper.GetLogger().Debugln("GetUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap Start")
	for s, _ := range d.seriesSubNeedDlMap {
		log_helper.GetLogger().Debugln(s)
	}
	log_helper.GetLogger().Debugln("GetUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap End")

	log_helper.GetLogger().Debugln("GetUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList Start")
	for s, value := range d.movieFileFullPathList {
		log_helper.GetLogger().Debugln(s, value)
	}
	log_helper.GetLogger().Debugln("GetUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList End")

	return nil
}

func (d Downloader) RefreshEmbySubList() error {

	if d.embyHelper == nil {
		return nil
	}

	bRefresh := false
	defer func() {
		if bRefresh == true {
			d.log.Infoln("Refresh Emby Sub List Success")
		} else {
			d.log.Errorln("Refresh Emby Sub List Error")
		}
	}()

	bRefresh, err := d.embyHelper.RefreshEmbySubList()
	if err != nil {
		return err
	}

	return nil
}

// DownloadSub4Movie 这里对接 Emby 的时候比较方便，只要更新 d.movieFileFullPathList 就行了，不像连续剧那么麻烦
func (d Downloader) DownloadSub4Movie(dir string) error {
	defer func() {
		// 所有的电影字幕下载完成，抉择完成，需要清理缓存目录
		err := my_util.ClearRootTmpFolder()
		if err != nil {
			d.log.Error("ClearRootTmpFolder", err)
		}
		d.log.Infoln("Download Movie Sub End...")
	}()
	var err error
	d.log.Infoln("Download Movie Sub Started...")
	// -----------------------------------------------------
	// 优先判断特殊的操作
	if d.needForcedScanAndDownSub == true {
		// 全扫描
		d.movieFileFullPathList, err = my_util.SearchMatchedVideoFile(dir)
		if err != nil {
			return err
		}
	} else {
		// 是否是通过 emby_helper api 获取的列表
		if d.embyHelper == nil {
			// 没有填写 emby_helper api 的信息，那么就走常规的全文件扫描流程
			d.movieFileFullPathList, err = my_util.SearchMatchedVideoFile(dir)
			if err != nil {
				return err
			}
		} else {
			// 进过 emby_helper api 的信息读取
			d.log.Infoln("Movie Sub Dl From Emby API...")
			if len(d.movieFileFullPathList) < 1 {
				d.log.Infoln("Movie Sub Dl From Emby API no movie need Dl sub")
				return nil
			}
		}
	}
	// -----------------------------------------------------
	// 并发控制
	movieDlFunc := func(i interface{}) error {
		inData := i.(InputData)
		// -----------------------------------------------------
		// 构建每个字幕站点下载者的实例
		var subSupplierHub = subSupplier.NewSubSupplierHub(
			//subhd.NewSupplier(d.reqParam),
			zimuku.NewSupplier(d.reqParam),
			xunlei.NewSupplier(d.reqParam),
			shooter.NewSupplier(d.reqParam),
		)

		if common.SubhdCode != "" {
			// 如果找到 code 了，那么就可以继续用这个实例
			subSupplierHub.AddSubSupplier(subhd.NewSupplier(d.reqParam))
		}
		// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
		organizeSubFiles, err := subSupplierHub.DownloadSub4Movie(inData.OneVideoFullPath, inData.Index, d.needForcedScanAndDownSub)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Movie", inData.OneVideoFullPath, err)
			return err
		}
		// 返回的两个值都是 nil 的时候，就是无需下载字幕，那么同样不用输出额外的信息，因为之前会输出跳过的原因
		if organizeSubFiles == nil {
			return nil
		}
		// 去搜索了没有发现字幕
		if len(organizeSubFiles) < 1 {
			d.log.Infoln("no sub found", filepath.Base(inData.OneVideoFullPath))
			return nil
		}
		d.oneVideoSelectBestSub(inData.OneVideoFullPath, organizeSubFiles)
		// -----------------------------------------------------

		return nil
	}
	// -----------------------------------------------------
	antPool, err := ants.NewPoolWithFunc(d.reqParam.Threads, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), common.OneMovieProcessTimeOut)
		defer cancel()

		done := make(chan error, 1)
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			done <- movieDlFunc(inData)
		}()

		select {
		case err := <-done:
			if err != nil {
				d.log.Errorln("DownloadSub4Movie.NewPoolWithFunc done with Error", err.Error())
			}
			return
		case p := <-panicChan:
			d.log.Errorln("DownloadSub4Movie.NewPoolWithFunc got panic", p)
			return
		case <-ctx.Done():
			d.log.Errorln("DownloadSub4Movie.NewPoolWithFunc got time out", ctx.Err())
			return
		}
	})
	if err != nil {
		return err
	}
	defer antPool.Release()
	wg := sync.WaitGroup{}
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range d.movieFileFullPathList {
		wg.Add(1)
		err = antPool.Invoke(InputData{OneVideoFullPath: oneVideoFullPath, Index: i, Wg: &wg})
		if err != nil {
			d.log.Errorln("DownloadSub4Movie ants.Invoke", err)
		}
	}
	wg.Wait()
	return nil
}

func (d Downloader) DownloadSub4Series(dir string) error {
	var err error
	defer func() {
		// 所有的连续剧字幕下载完成，抉择完成，需要清理缓存目录
		err := my_util.ClearRootTmpFolder()
		if err != nil {
			d.log.Error("ClearRootTmpFolder", err)
		}
		d.log.Infoln("Download Series Sub End...")

		my_util.CloseChrome()
		d.log.Infoln("CloseChrome")
	}()
	d.log.Infoln("Download Series Sub Started...")

	// 并发控制
	seriesDlFunc := func(i interface{}) error {
		inData := i.(InputData)
		// 构建每个字幕站点下载者的实例
		var subSupplierHub *subSupplier.SubSupplierHub
		subSupplierHub = subSupplier.NewSubSupplierHub(
			zimuku.NewSupplier(d.reqParam),
			//subhd.NewSupplier(d.reqParam),
			xunlei.NewSupplier(d.reqParam),
			shooter.NewSupplier(d.reqParam),
		)
		// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
		var seriesInfo *series.SeriesInfo
		var organizeSubFiles map[string][]string
		// 优先判断特殊的操作
		if d.needForcedScanAndDownSub == true {
			// 全盘扫描
			seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4Series(inData.OneVideoFullPath, inData.Index, d.needForcedScanAndDownSub)
			if err != nil {
				d.log.Errorln("subSupplierHub.DownloadSub4Series", inData.OneVideoFullPath, err)
				return err
			}
		} else {
			// 是否是通过 emby_helper api 获取的列表
			if d.embyHelper == nil {
				seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4Series(inData.OneVideoFullPath, inData.Index, d.needForcedScanAndDownSub)
				if err != nil {
					d.log.Errorln("subSupplierHub.DownloadSub4Series", inData.OneVideoFullPath, err)
					return err
				}
			} else {
				// 先进性 emby_helper api 的操作，读取需要更新字幕的项目
				seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4SeriesFromEmby(
					filepath.Join(dir, inData.OneVideoFullPath),
					d.seriesSubNeedDlMap[inData.OneVideoFullPath], inData.Index)
				if err != nil {
					d.log.Errorln("subSupplierHub.DownloadSub4Series", inData.OneVideoFullPath, err)
					return err
				}
			}
		}
		if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
			d.log.Infoln("no sub found", filepath.Base(inData.OneVideoFullPath))
			return nil
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
			seasonEpsKey := my_util.GetEpisodeKeyName(episodeInfo.Season, episodeInfo.Episode)
			d.oneVideoSelectBestSub(episodeInfo.FileFullPath, fullSeasonSubDict[seasonEpsKey])
		}
		// 是否清理全季的缓存字幕文件夹
		if d.reqParam.SaveOneSeasonSub == false {
			err = sub_helper.DeleteOneSeasonSubCacheFolder(seriesInfo.DirPath)
			if err != nil {
				return err
			}
		}

		return nil
	}
	antPool, err := ants.NewPoolWithFunc(d.reqParam.Threads, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), common.OneSeriesProcessTimeOut)
		defer cancel()

		done := make(chan error, 1)
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			done <- seriesDlFunc(inData)
		}()

		select {
		case err := <-done:
			if err != nil {
				d.log.Errorln("DownloadSub4Series.NewPoolWithFunc done with Error", err.Error())
			}
			return
		case p := <-panicChan:
			d.log.Errorln("DownloadSub4Series.NewPoolWithFunc got panic", p)
		case <-ctx.Done():
			d.log.Errorln("DownloadSub4Series.NewPoolWithFunc got time out", ctx.Err())
			return
		}
	})
	if err != nil {
		return err
	}
	defer antPool.Release()

	// 是否是通过 emby_helper api 获取的列表
	var seriesDirList = make([]string, 0)
	if d.embyHelper == nil {
		// 遍历连续剧总目录下的第一层目录
		seriesDirList, err = seriesHelper.GetSeriesList(dir)
		if err != nil {
			return err
		}
		for index, seriesDir := range seriesDirList {
			d.log.Debugln("embyHelper == nil GetSeriesList", index, seriesDir)
		}
	} else {
		// 这里给出的是连续剧的文件夹名称
		d.log.Debugln("embyHelper seriesSubNeedDlMap Count:", len(d.seriesSubNeedDlMap))
		for s, _ := range d.seriesSubNeedDlMap {
			seriesDirList = append(seriesDirList, s)
			d.log.Debugln("embyHelper seriesSubNeedDlMap:", s)
		}
	}
	wg := sync.WaitGroup{}
	for i, oneSeriesPath := range seriesDirList {
		wg.Add(1)
		err = antPool.Invoke(InputData{OneVideoFullPath: oneSeriesPath, Index: i, Wg: &wg})
		if err != nil {
			d.log.Errorln("DownloadSub4Series ants.Invoke", err)
		}
	}
	wg.Wait()
	return nil
}

func (d Downloader) RestoreFixTimelineBK(moviesDir, seriesDir string) error {

	defer d.log.Infoln("End Restore Fix Timeline BK")
	d.log.Infoln("Start Restore Fix Timeline BK...")
	_, err := sub_timeline_fixer_pkg.Restore(moviesDir, seriesDir)
	if err != nil {
		return err
	}
	return nil
}

// oneVideoSelectBestSub 一个视频，选择最佳的一个字幕（也可以保存所有网站第一个最佳字幕）
func (d Downloader) oneVideoSelectBestSub(oneVideoFullPath string, organizeSubFiles []string) {

	// 如果没有则直接跳过
	if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
		return
	}

	var err error
	// 得到目标视频文件的文件名
	videoFileName := filepath.Base(oneVideoFullPath)
	// -------------------------------------------------
	// 调试缓存，把下载好的字幕写到对应的视频目录下，方便调试
	if d.reqParam.DebugMode == true {

		err = my_util.CopyFiles2DebugFolder([]string{videoFileName}, organizeSubFiles)
		if err != nil {
			d.log.Errorln("copySubFile2DesFolder", err)
		}
	}
	// -------------------------------------------------
	/*
		这里需要额外考虑一点，有可能当前目录已经有一个 .Default .Forced 标记的字幕了
		那么下载字幕丢进来的时候就需要提前把这个字幕找出来，去除整个 .Default .Forced  标记
		然后进行正常的下载，存储和替换字幕，最后将本次操作的第一次标记为 .Default
	*/
	// 不管是不是保存多个字幕，都要先扫描本地的字幕，进行 .Default .Forced 去除
	// 这个视频的所有字幕，去除 .default .Forced 标记
	err = sub_helper.SearchVideoMatchSubFileAndRemoveExtMark(oneVideoFullPath)
	if err != nil {
		// 找个错误可以忍
		d.log.Errorln("SearchVideoMatchSubFileAndRemoveExtMark,", oneVideoFullPath, err)
	}
	if d.reqParam.SaveMultiSub == false {
		// 选择最优的一个字幕
		var finalSubFile *subparser.FileInfo
		finalSubFile = d.mk.SelectOneSubFile(organizeSubFiles)
		if finalSubFile == nil {
			d.log.Warnln("Found", len(organizeSubFiles), " subtitles but not one fit:", oneVideoFullPath)
			return
		}
		/*
			这里还有一个梗，Emby、jellyfin 支持 default 和 forced 扩展字段
			但是，plex 只支持 forced
			那么就比较麻烦，干脆，normal 的命名格式化实例，就不设置 default 了，forced 不想用，因为可能会跟你手动选择的字幕冲突（下次观看的时候，理论上也可能不会）
		*/
		// 判断配置文件中的字幕命名格式化的选择
		bSetDefault := true
		if d.subNameFormatter == subcommon.Normal {
			bSetDefault = false
		}
		// 找到了，写入文件
		err = d.writeSubFile2VideoPath(oneVideoFullPath, *finalSubFile, "", bSetDefault, false)
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
		// 多网站 Top 1 字幕保存的时候，第一个设置为 Default 即可
		/*
			由于新功能支持了字幕命名格式的选择，那么如果触发了多个字幕保存的逻辑，如果不调整
			则会遇到，top1 先写入，然后 top2 覆盖 top1 ，以此类推的情况出现
			所以如果开启了 Normal SubNameFormatter 的功能，则要反序写入文件
			如果是 Emby 的字幕命名格式则无需考虑此问题，因为每个网站只会有一个字幕，且字幕命名格式决定了不会重复写入覆盖
		*/
		if d.subNameFormatter == subcommon.Emby {
			for i, file := range finalSubFiles {
				setDefault := false
				if i == 0 {
					setDefault = true
				}
				err = d.writeSubFile2VideoPath(oneVideoFullPath, file, siteNames[i], setDefault, false)
				if err != nil {
					d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub, "writeSubFile2VideoPath:", err)
					return
				}
			}
		} else {
			// 默认这里就是 normal 模式
			// 逆序写入
			/*
				这里还有一个梗，Emby、jellyfin 支持 default 和 forced 扩展字段
				但是，plex 只支持 forced
				那么就比较麻烦，干脆，normal 的命名格式化实例，就不设置 default 了，forced 不想用，因为可能会跟你手动选择的字幕冲突（下次观看的时候，理论上也可能不会）
			*/
			for i := len(finalSubFiles) - 1; i > -1; i-- {
				err = d.writeSubFile2VideoPath(oneVideoFullPath, finalSubFiles[i], siteNames[i], false, false)
				if err != nil {
					d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub, "writeSubFile2VideoPath:", err)
					return
				}
			}
		}
	}
	// -------------------------------------------------
}

// saveFullSeasonSub 这里就需要单独存储到连续剧每一季的文件夹的特殊文件夹中。需要跟 DeleteOneSeasonSubCacheFolder 关联起来
func (d Downloader) saveFullSeasonSub(seriesInfo *series.SeriesInfo, organizeSubFiles map[string][]string) map[string][]string {

	var fullSeasonSubDict = make(map[string][]string)

	for _, season := range seriesInfo.SeasonDict {
		seasonKey := my_util.GetEpisodeKeyName(season, 0)
		subs, ok := organizeSubFiles[seasonKey]
		if ok == false {
			continue
		}
		for _, sub := range subs {
			subFileName := filepath.Base(sub)

			newSeasonSubRootPath, err := my_util.GetDebugFolderByName([]string{
				filepath.Base(seriesInfo.DirPath),
				"Sub_" + seasonKey})
			if err != nil {
				d.log.Errorln("saveFullSeasonSub.GetDebugFolderByName", subFileName, err)
				continue
			}

			newSubFullPath := filepath.Join(newSeasonSubRootPath, subFileName)
			err = my_util.CopyFile(sub, newSubFullPath)
			if err != nil {
				d.log.Errorln("saveFullSeasonSub.CopyFile", subFileName, err)
				continue
			}
			// 从字幕的文件名推断是 哪一季 的 那一集
			_, gusSeason, gusEpisode, err := decode.GetSeasonAndEpisodeFromSubFileName(subFileName)
			if err != nil {
				return nil
			}
			// 把整季的字幕缓存位置也提供出去，如果之前没有下载到的，这里返回出来的可以补上
			seasonEpsKey := my_util.GetEpisodeKeyName(gusSeason, gusEpisode)
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

// 在前面需要进行语言的筛选、排序，这里仅仅是存储， extraSubPreName 这里传递是字幕的网站，有就认为是多字幕的存储。空就是单字幕，单字幕就可以setDefault
func (d Downloader) writeSubFile2VideoPath(videoFileFullPath string, finalSubFile subparser.FileInfo, extraSubPreName string, setDefault bool, skipExistFile bool) error {
	defer d.log.Infoln("----------------------------------")
	videoRootPath := filepath.Dir(videoFileFullPath)
	subNewName, subNewNameWithDefault, _ := d.subFormatter.GenerateMixSubName(videoFileFullPath, finalSubFile.Ext, finalSubFile.Lang, extraSubPreName)

	desSubFullPath := filepath.Join(videoRootPath, subNewName)
	if setDefault == true {
		// 先判断没有 default 的字幕是否存在了，在的话，先删除，然后再写入
		if my_util.IsFile(desSubFullPath) == true {
			_ = os.Remove(desSubFullPath)
		}
		desSubFullPath = filepath.Join(videoRootPath, subNewNameWithDefault)
	}

	if skipExistFile == true {
		// 需要判断文件是否存在在，有则跳过
		if my_util.IsFile(desSubFullPath) == true {
			d.log.Infoln("OrgSubName:", finalSubFile.Name)
			d.log.Infoln("Sub Skip DownAt:", desSubFullPath)
			return nil
		}
	}
	// 最后写入字幕
	err := my_util.WriteFile(desSubFullPath, finalSubFile.Data)
	if err != nil {
		return err
	}
	d.log.Infoln("----------------------------------")
	d.log.Infoln("OrgSubName:", finalSubFile.Name)
	d.log.Infoln("SubDownAt:", desSubFullPath)

	// 然后还需要判断是否需要校正字幕的时间轴
	if d.reqParam.FixTimeLine == true {
		err = d.subTimelineFixerHelperEx.Process(videoFileFullPath, desSubFullPath)
		if err != nil {
			return err
		}
	}

	return nil
}
