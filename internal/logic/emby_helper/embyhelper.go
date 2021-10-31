package emby_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/emby_api"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/net/context"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type EmbyHelper struct {
	embyApi    *embyHelper.EmbyApi
	EmbyConfig emby.EmbyConfig
	threads    int
	timeOut    time.Duration
	listLock   sync.Mutex
}

func NewEmbyHelper(embyConfig emby.EmbyConfig) *EmbyHelper {
	em := EmbyHelper{EmbyConfig: embyConfig}
	em.embyApi = embyHelper.NewEmbyApi(embyConfig)
	em.threads = 6
	em.timeOut = 60 * time.Second
	return &em
}

func (em *EmbyHelper) GetRecentlyAddVideoList(movieRootDir, seriesRootDir string) ([]emby.EmbyMixInfo, map[string][]emby.EmbyMixInfo, error) {

	// 获取电影和连续剧的文件夹名称
	movieFolderName := filepath.Base(movieRootDir)
	seriesFolderName := filepath.Base(seriesRootDir)

	var EpisodeIdList = make([]string, 0)
	var MovieIdList = make([]string, 0)
	// 获取最近的影片列表
	items, err := em.embyApi.GetRecentlyItems()
	if err != nil {
		return nil, nil, err
	}

	log_helper.GetLogger().Debugln("GetRecentlyAddVideoList - GetRecentlyItems Count", len(items.Items))

	// 分类
	for _, item := range items.Items {
		if item.Type == videoTypeEpisode {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			EpisodeIdList = append(EpisodeIdList, item.Id)
		} else if item.Type == videoTypeMovie {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			MovieIdList = append(MovieIdList, item.Id)
		} else {
			log_helper.GetLogger().Debugln("GetRecentlyItems - Is not a goal video type:", item.Type, item.Name)
		}
	}
	// 过滤出有效的电影、连续剧的资源出来
	filterMovieList, err := em.filterEmbyVideoList(movieFolderName, MovieIdList, true)
	if err != nil {
		return nil, nil, err
	}
	filterSeriesList, err := em.filterEmbyVideoList(seriesFolderName, EpisodeIdList, false)
	if err != nil {
		return nil, nil, err
	}
	// 将没有字幕的找出来
	noSubMovieList, err := em.filterNoChineseSubVideoList(filterMovieList)
	if err != nil {
		return nil, nil, err
	}
	noSubSeriesList, err := em.filterNoChineseSubVideoList(filterSeriesList)
	if err != nil {
		return nil, nil, err
	}
	// 拼接绝对路径
	for i, info := range noSubMovieList {
		noSubMovieList[i].VideoFileFullPath = filepath.Join(movieRootDir, info.VideoFileRelativePath)
	}
	for i, info := range noSubSeriesList {
		noSubSeriesList[i].VideoFileFullPath = filepath.Join(seriesRootDir, info.VideoFileRelativePath)
	}
	// 需要将连续剧零散的每一集，进行合并到一个连续剧下面，也就是这个连续剧有那些需要更新的
	var seriesMap = make(map[string][]emby.EmbyMixInfo)
	for _, info := range noSubSeriesList {
		_, ok := seriesMap[info.VideoFolderName]
		if ok == false {
			// 不存在则新建初始化
			seriesMap[info.VideoFolderName] = make([]emby.EmbyMixInfo, 0)
		}
		seriesMap[info.VideoFolderName] = append(seriesMap[info.VideoFolderName], info)
	}

	return noSubMovieList, seriesMap, nil
}

// RefreshEmbySubList 字幕下载完毕一次，就可以触发一次这个。并发 6 线程去刷新
func (em *EmbyHelper) RefreshEmbySubList() (bool, error) {
	if em.embyApi == nil {
		return false, nil
	}
	err := em.embyApi.RefreshRecentlyVideoInfo()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (em *EmbyHelper) filterEmbyVideoList(videoFolderName string, videoIdList []string, isMovieOrSeries bool) ([]emby.EmbyMixInfo, error) {
	var filterVideoEmbyInfo = make([]emby.EmbyMixInfo, 0)

	queryFunc := func(m string) (*emby.EmbyMixInfo, error) {
		info, err := em.embyApi.GetItemVideoInfo(m)
		if err != nil {
			return nil, err
		}
		ancs, err := em.embyApi.GetItemAncestors(m)
		if err != nil {
			return nil, err
		}
		mixInfo := emby.EmbyMixInfo{Ancestors: ancs, VideoInfo: info}
		if isMovieOrSeries == true {
			// 电影
			// 过滤掉不符合要求的
			if len(mixInfo.Ancestors) < 2 {
				return nil, err
			}
			// 过滤掉不符合要求的
			if mixInfo.Ancestors[0].Name != videoFolderName || mixInfo.Ancestors[0].Type != "Folder" {
				return nil, err
			}
			// 这个电影的文件夹
			mixInfo.VideoFolderName = filepath.Base(filepath.Dir(mixInfo.VideoInfo.Path))
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
			mixInfo.VideoFileRelativePath = filepath.Join(mixInfo.VideoFolderName, mixInfo.VideoFileName)
		} else {
			// 连续剧
			// 过滤掉不符合要求的
			if len(mixInfo.Ancestors) < 3 {
				return nil, err
			}
			// 过滤掉不符合要求的
			if mixInfo.Ancestors[0].Type != "Season" ||
				mixInfo.Ancestors[1].Type != "Series" ||
				mixInfo.Ancestors[2].Type != "Folder" ||
				mixInfo.Ancestors[2].Name != videoFolderName {
				return nil, err
			}
			// 这个剧集的文件夹
			mixInfo.VideoFolderName = filepath.Base(mixInfo.Ancestors[1].Path)
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
			seasonName := filepath.Base(mixInfo.Ancestors[0].Path)
			mixInfo.VideoFileRelativePath = filepath.Join(mixInfo.VideoFolderName, seasonName, mixInfo.VideoFileName)
		}

		return &mixInfo, nil
	}

	p, err := ants.NewPoolWithFunc(em.threads, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), em.timeOut)
		defer cancel()

		done := make(chan OutData, 1)
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			info, err := queryFunc(data.Id)
			outData := OutData{
				Info: info,
				Err:  err,
			}
			done <- outData
		}()

		select {
		case outData := <-done:
			// 收到结果，需要加锁
			if outData.Err != nil {
				log_helper.GetLogger().Errorln("filterEmbyVideoList.NewPoolWithFunc got Err", outData.Err)
				return
			}
			if outData.Info == nil {
				return
			}
			em.listLock.Lock()
			filterVideoEmbyInfo = append(filterVideoEmbyInfo, *outData.Info)
			em.listLock.Unlock()
			return
		case p := <-panicChan:
			log_helper.GetLogger().Errorln("filterEmbyVideoList.NewPoolWithFunc got panic", p)
		case <-ctx.Done():
			log_helper.GetLogger().Errorln("filterEmbyVideoList.NewPoolWithFunc got time out", ctx.Err())
			return
		}
	})
	if err != nil {
		return nil, err
	}
	defer p.Release()
	wg := sync.WaitGroup{}
	// 获取视频的 Emby 信息
	for _, m := range videoIdList {
		wg.Add(1)
		err = p.Invoke(InputData{Id: m, Wg: &wg})
		if err != nil {
			log_helper.GetLogger().Errorln("filterEmbyVideoList ants.Invoke", err)
		}
	}
	wg.Wait()

	return filterVideoEmbyInfo, nil
}

func (em *EmbyHelper) filterNoChineseSubVideoList(videoList []emby.EmbyMixInfo) ([]emby.EmbyMixInfo, error) {
	currentTime := time.Now()
	dayRange3Months, _ := time.ParseDuration(common.DownloadSubDuring3Months)
	dayRange7Days, _ := time.ParseDuration(common.DownloadSubDuring7Days)

	var noSubVideoList = make([]emby.EmbyMixInfo, 0)
	// TODO 这里有一种情况需要考虑的，如果内置有中文的字幕，那么是否需要跳过，目前暂定的一定要有外置的字幕
	for _, info := range videoList {

		needDlSub3Month := false
		// 3个月内，或者没有字幕都要进行下载
		if info.VideoInfo.PremiereDate.Add(dayRange3Months).After(currentTime) == true {
			// 需要下载的
			needDlSub3Month = true
		}
		// 这个影片只要有一个符合字幕要求的，就可以跳过
		// 外置中文字幕
		haveExternalChineseSub := false
		for _, stream := range info.VideoInfo.MediaStreams {
			// 首先找到外置的字幕文件
			if stream.IsExternal == true && stream.IsTextSubtitleStream == true && stream.SupportsExternalStream == true {
				// 然后字幕的格式以及语言命名要符合本程序的定义，有字幕
				if em.subTypeStringOK(stream.Codec) == true && em.langStringOK(stream.Language) == true {
					haveExternalChineseSub = true
					break
				} else {
					continue
				}
			}
		}
		// 内置中文字幕
		haveInsideChineseSub := false
		for _, stream := range info.VideoInfo.MediaStreams {
			if stream.IsExternal == false && (stream.Language == "chi" || stream.Language == "cht" || stream.Language == "chs") {
				haveInsideChineseSub = true
				break
			}
		}
		// 比如，创建的时间在3个月内，然后没有额外下载的中文字幕，都符合要求
		if haveExternalChineseSub == false {
			// 如果创建了7天，且有内置的中文字幕，那么也不进行下载了
			if info.VideoInfo.DateCreated.Add(dayRange7Days).After(currentTime) == false && haveInsideChineseSub == true {
				continue
			}
			//// 如果创建了三个月，还是没有字幕，那么也不进行下载了
			//if info.VideoInfo.DateCreated.Add(dayRange3Months).After(currentTime) == false {
			//	continue
			//}
			// 没有中文字幕就加入下载列表
			noSubVideoList = append(noSubVideoList, info)
		} else {
			// 如果视频发布时间超过两年了，有字幕就直接跳过了，一般字幕稳定了
			if currentTime.Year()-2 > info.VideoInfo.PremiereDate.Year() {
				continue
			}
			// 有中文字幕，且如果在三个月内，则需要继续下载字幕`
			if needDlSub3Month == true {
				noSubVideoList = append(noSubVideoList, info)
			}
		}
	}

	return noSubVideoList, nil
}

// GetInternalEngSubAndExChineseEnglishSub 获取对应 videoId 的内置英文字幕，外置中（简体、繁体）英字幕
func (em *EmbyHelper) GetInternalEngSubAndExChineseEnglishSub(videoId string) (bool, []emby.SubInfo, []emby.SubInfo, error) {

	// 先刷新以下这个资源，避免找到的字幕不存在了
	err := em.embyApi.UpdateVideoSubList(videoId)
	if err != nil {
		return false, nil, nil, err
	}
	// 获取这个资源的信息
	videoInfo, err := em.embyApi.GetItemVideoInfo(videoId)
	if err != nil {
		return false, nil, nil, err
	}
	// 获取 MediaSources ID，这里强制使用第一个视频源（因为 emby 运行有多个版本的视频指向到一个视频ID上，比如一个 web 一个 蓝光）
	mediaSourcesId := videoInfo.MediaSources[0].Id
	// 视频文件名称带后缀名
	videoFileName := filepath.Base(videoInfo.Path)
	videoFileNameWithOutExt := strings.ReplaceAll(videoFileName, path.Ext(videoFileName), "")
	// TODO 后续会新增一个功能，从视频中提取音频文件，然后识别转为字符，再进行与字幕的匹配
	// 获取是否有内置的英文字幕，如果没有则无需继续往下
	/*
		这里有个梗，读取到的英文内置字幕很可能是残缺的，比如，基地 S01E04 Eng 第一个 Default Forced Sub，就不对，内容的 Dialogue 很少。
		然后第二个 Eng 字幕才对。那么考虑到兼容性， 可能后续有短视频，也就不能简单的按 Dialogue 的多少去衡量。大概会做一个功能。方案有两个：
		1. 读取到视频的总长度，然后再分析 Dialogue 的时间出现的部分与整体时间轴的占比，又或者是 Dialogue 之间的连续成都分析，这个有待测试。
		2. 还有一个更加粗暴的方案，把所有的 Eng 都识别出来，然后找最多的 Dialogue 来做为正确的来使用（够粗暴吧）
	*/
	var insideEngSUbIndexList = make([]int, 0)
	for _, stream := range videoInfo.MediaStreams {
		if stream.IsExternal == false && stream.Language == "eng" && stream.Codec == "subrip" {
			insideEngSUbIndexList = append(insideEngSUbIndexList, stream.Index)
		}
	}
	// 没有找到则跳过
	if len(insideEngSUbIndexList) == 0 {
		return false, nil, nil, nil
	}
	// 再内置英文字幕能找到的前提下，就可以先找中文的外置字幕，目前版本只能考虑双语字幕
	// 内置英文字幕，这里把 srt 和 ass 的都导出来
	var inSubList = make([]emby.SubInfo, 0)
	// 外置中文双语字幕
	var exSubList = make([]emby.SubInfo, 0)
	tmpFileNameWithOutExt := ""
	for _, stream := range videoInfo.MediaStreams {
		// 首先找到外置的字幕文件
		if stream.IsExternal == true && stream.IsTextSubtitleStream == true && stream.SupportsExternalStream == true {
			// 然后字幕的格式以及语言命名要符合本程序的定义，有字幕
			if em.subTypeStringOK(stream.Codec) == true &&
				em.langStringOK(stream.Language) == true &&
				// 只支持 简英、繁英
				(strings.Contains(stream.Language, language.MatchLangChsEn) == true || strings.Contains(stream.Language, language.MatchLangChtEn) == true) {

				tmpFileName := filepath.Base(stream.Path)
				// 去除 .default 或者 .forced
				tmpFileName = strings.ReplaceAll(tmpFileName, language.Sub_Ext_Mark_Default, "")
				tmpFileName = strings.ReplaceAll(tmpFileName, language.Sub_Ext_Mark_Forced, "")
				tmpFileNameWithOutExt = strings.ReplaceAll(tmpFileName, path.Ext(tmpFileName), "")
				exSubList = append(exSubList, *emby.NewSubInfo(tmpFileNameWithOutExt+"."+stream.Codec, "."+stream.Codec, stream.Index))
			} else {
				continue
			}
		}
	}
	// 没有找到则跳过
	if len(exSubList) == 0 {
		return false, nil, nil, nil
	}
	/*
		把之前 Internal 英文字幕的 SubInfo 实例的信息补充完整
		但是也不是绝对的，因为后续去 emby 下载字幕的时候，需要与外置字幕的后缀名一致
		这里开始去下载字幕
		先下载内置的文的
		因为上面下载内置英文字幕的梗，所以，需要预先下载多个内置的英文字幕下来，用体积最大（相同后缀名）的那个来作为最后的输出
	*/
	// 那么现在先下载相同格式（.srt）的两个字幕
	InsideEngSubIndex := 0
	if len(insideEngSUbIndexList) == 1 {
		// 如果就找到一个内置字幕，就默认这个
		InsideEngSubIndex = insideEngSUbIndexList[0]
	} else {
		// 如果找到不止一个就需要判断
		var tmpSubContentLenList = make([]int, 0)
		for _, index := range insideEngSUbIndexList {
			// TODO 这里默认是去 Emby 去拿字幕，但是其实可以缓存在视频文件同级的目录下，这样后续就无需多次下载了，毕竟每次下载都需要读取完整的视频
			subFileData, err := em.embyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", index), common.SubExtSRT)
			if err != nil {
				return false, nil, nil, err
			}
			tmpSubContentLenList = append(tmpSubContentLenList, len(subFileData))
		}
		maxContentLen := -1
		for index, contentLen := range tmpSubContentLenList {
			if maxContentLen < contentLen {
				maxContentLen = contentLen
				InsideEngSubIndex = insideEngSUbIndexList[index]
			}
		}
	}
	// 这里才是下载最佳的那个字幕
	for i := 0; i < 2; i++ {
		tmpExt := common.SubExtSRT
		if i == 1 {
			tmpExt = common.SubExtASS
		}
		subFileData, err := em.embyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", InsideEngSubIndex), tmpExt)
		if err != nil {
			return false, nil, nil, err
		}
		tmpInSubInfo := emby.NewSubInfo(videoFileNameWithOutExt+tmpExt, tmpExt, InsideEngSubIndex)
		tmpInSubInfo.Content = []byte(subFileData)
		inSubList = append(inSubList, *tmpInSubInfo)
	}
	// 再下载外置的
	for i, subInfo := range exSubList {
		subFileData, err := em.embyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", subInfo.EmbyStreamIndex), subInfo.Ext)
		if err != nil {
			return false, nil, nil, err
		}
		exSubList[i].Content = []byte(subFileData)
	}

	return true, inSubList, exSubList, nil
}

// langStringOK 从 Emby api 拿到字幕的 MyLanguage string是否是符合本程序要求的
func (em *EmbyHelper) langStringOK(inLang string) bool {

	tmpString := strings.ToLower(inLang)
	nextString := tmpString
	// 去除 [xunlie] 类似的标记
	spStrings := strings.Split(tmpString, "[")
	if len(spStrings) > 1 {
		nextString = spStrings[0]
	} else {
		spStrings = strings.Split(tmpString, "(")
		if len(spStrings) > 1 {
			nextString = spStrings[0]
		}
	}
	switch nextString {
	// 早期版本支持的语言类型，现在弃用
	case em.replaceLangString(language.Emby_chi),
		em.replaceLangString(language.Emby_chn),
		em.replaceLangString(language.Emby_chs),
		em.replaceLangString(language.Emby_cht),
		em.replaceLangString(language.Emby_chs_en),
		em.replaceLangString(language.Emby_cht_en),
		em.replaceLangString(language.Emby_chs_jp),
		em.replaceLangString(language.Emby_cht_jp),
		em.replaceLangString(language.Emby_chs_kr),
		em.replaceLangString(language.Emby_cht_kr):
		return true
	case em.replaceLangString(language.Emby_chinese):
		return true
	default:
		return false
	}
}

// subTypeStringOK 从 Emby api 拿到字幕的 sub 类型 string (Codec) 是否是符合本程序要求的
func (em *EmbyHelper) subTypeStringOK(inSubType string) bool {

	tmpString := strings.ToLower(inSubType)
	if tmpString == common.SubTypeSRT ||
		tmpString == common.SubTypeASS ||
		tmpString == common.SubTypeSSA {
		return true
	}

	return false
}

func (em *EmbyHelper) replaceLangString(inString string) string {
	tmpString := strings.ToLower(inString)
	one := strings.ReplaceAll(tmpString, ".", "")
	two := strings.ReplaceAll(one, "_", "")
	return two
}

type InputData struct {
	Id string
	Wg *sync.WaitGroup
}

type OutData struct {
	Info *emby.EmbyMixInfo
	Err  error
}

const (
	videoTypeEpisode = "Episode"
	videoTypeMovie   = "Movie"
)
