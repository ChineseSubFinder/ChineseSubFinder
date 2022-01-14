package downloader

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
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	subcommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	sub_timeline_fixer_pkg "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_control"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"path/filepath"
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

	taskControl *task_control.TaskControl
}

func NewDownloader(inSubFormatter ifaces.ISubFormatter, _reqParam ...types.ReqParam) (*Downloader, error) {

	var downloader Downloader
	var err error
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
	// 初始化任务控制
	downloader.taskControl, err = task_control.NewTaskControl(downloader.reqParam.Threads, log_helper.GetLogger())
	if err != nil {
		return nil, err
	}

	return &downloader, nil
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
	// 并发控制，设置为 movie 的处理函数
	d.taskControl.SetCtxProcessFunc("MoviePool", d.movieDlFunc, common.OneMovieProcessTimeOut)
	// -----------------------------------------------------
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range d.movieFileFullPathList {

		err = d.taskControl.Invoke(&task_control.TaskData{
			Index: i,
			DataEx: DownloadInputData{
				OneVideoFullPath: oneVideoFullPath,
			},
		})
		if err != nil {
			d.log.Errorln("DownloadSub4Movie Invoke Index:", i, "Error", err)
		}
	}

	d.taskControl.Hold()
	// 可以得到执行结果的统计信息
	successList, noExecuteList, errorList := d.taskControl.GetExecuteInfo()

	d.log.Infoln("--------------------------------------")
	d.log.Infoln("successList", len(successList))
	for i, indexId := range successList {
		d.log.Infoln(i, d.movieFileFullPathList[indexId])
	}
	d.log.Infoln("--------------------------------------")
	d.log.Infoln("noExecuteList", len(noExecuteList))
	for i, indexId := range noExecuteList {
		d.log.Infoln(i, d.movieFileFullPathList[indexId])
	}
	d.log.Infoln("--------------------------------------")
	d.log.Infoln("errorList", len(errorList))
	for i, indexId := range errorList {
		d.log.Infoln(i, d.movieFileFullPathList[indexId])
	}
	d.log.Infoln("--------------------------------------")

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
	// -----------------------------------------------------
	// 并发控制，设置为 movie 的处理函数
	d.taskControl.SetCtxProcessFunc("SeriesPool", d.seriesDlFunc, common.OneSeriesProcessTimeOut)
	// -----------------------------------------------------
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
		for s := range d.seriesSubNeedDlMap {
			seriesDirList = append(seriesDirList, s)
			d.log.Debugln("embyHelper seriesSubNeedDlMap:", s)
		}
	}

	for i, oneSeriesPath := range seriesDirList {

		err = d.taskControl.Invoke(&task_control.TaskData{
			Index: i,
			DataEx: DownloadInputData{
				RootDirPath:   dir,
				OneSeriesPath: oneSeriesPath,
			},
		})
		if err != nil {
			d.log.Errorln("DownloadSub4Movie Invoke Index:", i, "Error", err)
		}
	}

	d.taskControl.Hold()
	// 可以得到执行结果的统计信息
	successList, noExecuteList, errorList := d.taskControl.GetExecuteInfo()

	d.log.Infoln("--------------------------------------")
	d.log.Infoln("successList", len(successList))
	for i, indexId := range successList {
		d.log.Infoln(i, seriesDirList[indexId])
	}
	d.log.Infoln("--------------------------------------")
	d.log.Infoln("noExecuteList", len(noExecuteList))
	for i, indexId := range noExecuteList {
		d.log.Infoln(i, seriesDirList[indexId])
	}
	d.log.Infoln("--------------------------------------")
	d.log.Infoln("errorList", len(errorList))
	for i, indexId := range errorList {
		d.log.Infoln(i, seriesDirList[indexId])
	}
	d.log.Infoln("--------------------------------------")

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

func (d Downloader) movieDlFunc(ctx context.Context, inData interface{}) error {

	taskData := inData.(*task_control.TaskData)
	downloadInputData := taskData.DataEx.(DownloadInputData)
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
	organizeSubFiles, err := subSupplierHub.DownloadSub4Movie(downloadInputData.OneVideoFullPath, taskData.Index, d.needForcedScanAndDownSub)
	if err != nil {
		d.log.Errorln("subSupplierHub.DownloadSub4Movie", downloadInputData.OneVideoFullPath, err)
		return err
	}
	// 返回的两个值都是 nil 的时候，就是无需下载字幕，那么同样不用输出额外的信息，因为之前会输出跳过的原因
	if organizeSubFiles == nil {
		return nil
	}
	// 去搜索了没有发现字幕
	if len(organizeSubFiles) < 1 {
		d.log.Infoln("no sub found", filepath.Base(downloadInputData.OneVideoFullPath))
		return nil
	}
	d.oneVideoSelectBestSub(downloadInputData.OneVideoFullPath, organizeSubFiles)
	// -----------------------------------------------------

	return nil
}

func (d Downloader) seriesDlFunc(ctx context.Context, inData interface{}) error {

	var err error
	taskData := inData.(*task_control.TaskData)
	downloadInputData := taskData.DataEx.(DownloadInputData)
	// 构建每个字幕站点下载者的实例
	var subSupplierHub *subSupplier.SubSupplierHub
	subSupplierHub = subSupplier.NewSubSupplierHub(
		zimuku.NewSupplier(d.reqParam),
		//subhd.NewSupplier(d.reqParam),
		xunlei.NewSupplier(d.reqParam),
		shooter.NewSupplier(d.reqParam),
	)
	if common.SubhdCode != "" {
		// 如果找到 code 了，那么就可以继续用这个实例
		subSupplierHub.AddSubSupplier(subhd.NewSupplier(d.reqParam))
	}

	// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
	var seriesInfo *series.SeriesInfo
	var organizeSubFiles map[string][]string
	// 优先判断特殊的操作
	if d.needForcedScanAndDownSub == true {
		// 全盘扫描
		seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4Series(downloadInputData.OneSeriesPath, taskData.Index, d.needForcedScanAndDownSub)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub4Series", downloadInputData.OneSeriesPath, err)
			return err
		}
	} else {
		// 是否是通过 emby_helper api 获取的列表
		if d.embyHelper == nil {
			// 不适用 emby api
			seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4Series(downloadInputData.OneSeriesPath, taskData.Index, d.needForcedScanAndDownSub)
			if err != nil {
				d.log.Errorln("subSupplierHub.DownloadSub4Series", downloadInputData.OneSeriesPath, err)
				return err
			}
		} else {
			// 先进性 emby_helper api 的操作，读取需要更新字幕的项目
			seriesInfo, organizeSubFiles, err = subSupplierHub.DownloadSub4SeriesFromEmby(
				filepath.Join(downloadInputData.RootDirPath, downloadInputData.OneSeriesPath),
				d.seriesSubNeedDlMap[downloadInputData.OneSeriesPath], taskData.Index)
			if err != nil {
				d.log.Errorln("subSupplierHub.DownloadSub4Series", downloadInputData.OneSeriesPath, err)
				return err
			}
		}
	}
	if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
		d.log.Infoln("no sub found", filepath.Base(downloadInputData.OneSeriesPath))
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
