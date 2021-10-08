package emby_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
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
	em.embyApi = embyHelper.NewEmbyHelper(embyConfig)
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
	// 分类
	for _, item := range items.Items {
		if item.Type == "Episode" {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			EpisodeIdList = append(EpisodeIdList, item.Id)
		} else if item.Type == "Movie" {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			MovieIdList = append(MovieIdList, item.Id)
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
		noSubMovieList[i].VideoFileFullPath = path.Join(movieRootDir, info.VideoFileRelativePath)
	}
	for i, info := range noSubSeriesList {
		noSubSeriesList[i].VideoFileFullPath = path.Join(seriesRootDir, info.VideoFileRelativePath)
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
		// 外置字幕
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
		// 内置字幕
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

// langStringOK 从 Emby api 拿到字幕的 Language string是否是符合本程序要求的
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
	case em.replaceLangString(types.Emby_chi),
		em.replaceLangString(types.Emby_chn),
		em.replaceLangString(types.Emby_chs),
		em.replaceLangString(types.Emby_cht),
		em.replaceLangString(types.Emby_chs_en),
		em.replaceLangString(types.Emby_cht_en),
		em.replaceLangString(types.Emby_chs_jp),
		em.replaceLangString(types.Emby_cht_jp),
		em.replaceLangString(types.Emby_chs_kr),
		em.replaceLangString(types.Emby_cht_kr):
		return true
	case em.replaceLangString(types.Emby_chinese):
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
