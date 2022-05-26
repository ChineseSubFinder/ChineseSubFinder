package emby_helper

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/emby_api"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/path_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sort_things"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type EmbyHelper struct {
	EmbyApi  *embyHelper.EmbyApi
	log      *logrus.Logger
	settings *settings.Settings
	timeOut  time.Duration
	listLock sync.Mutex
}

func NewEmbyHelper(_log *logrus.Logger, _settings *settings.Settings) *EmbyHelper {
	em := EmbyHelper{log: _log, settings: _settings}
	em.EmbyApi = embyHelper.NewEmbyApi(_log, _settings.EmbySettings)
	em.timeOut = 60 * time.Second
	return &em
}

// GetRecentlyAddVideoListWithNoChineseSubtitle 获取最近新添加的视频，且没有中文字幕的
func (em *EmbyHelper) GetRecentlyAddVideoListWithNoChineseSubtitle(needForcedScanAndDownSub ...bool) ([]emby.EmbyMixInfo, map[string][]emby.EmbyMixInfo, error) {

	filterMovieList, filterSeriesList, err := em.GetRecentlyAddVideoList()
	if err != nil {
		return nil, nil, err
	}
	var noSubMovieList, noSubSeriesList []emby.EmbyMixInfo
	var seriesMap = make(map[string][]emby.EmbyMixInfo)

	if len(needForcedScanAndDownSub) > 0 && needForcedScanAndDownSub[0] == true {
		// 强制扫描，无需过滤
		// 输出调试信息
		em.log.Debugln("-----------------")
		em.log.Debugln("GetRecentlyAddVideoList movie", len(filterMovieList))
		for index, info := range filterMovieList {
			em.log.Debugln(index, info.VideoFileName)
		}
		em.log.Debugln("-----------------")
		em.log.Debugln("GetRecentlyAddVideoList series", len(filterSeriesList))
		for index, info := range filterSeriesList {
			em.log.Debugln(index, info.VideoFileName)
		}
		em.log.Debugln("-----------------")

		// 需要将连续剧零散的每一集，进行合并到一个连续剧下面，也就是这个连续剧有那些需要更新的
		for _, info := range filterSeriesList {
			_, ok := seriesMap[info.VideoFolderName]
			if ok == false {
				// 不存在则新建初始化
				seriesMap[info.VideoFolderName] = make([]emby.EmbyMixInfo, 0)
			}
			seriesMap[info.VideoFolderName] = append(seriesMap[info.VideoFolderName], info)
		}

		return filterMovieList, seriesMap, nil

	} else {
		// 将没有字幕的找出来
		noSubMovieList, err = em.filterNoChineseSubVideoList(filterMovieList)
		if err != nil {
			return nil, nil, err
		}
		em.log.Debugln("-----------------")
		noSubSeriesList, err = em.filterNoChineseSubVideoList(filterSeriesList)
		if err != nil {
			return nil, nil, err
		}
		// 输出调试信息
		em.log.Debugln("-----------------")
		em.log.Debugln("filterNoChineseSubVideoList found no chinese movie", len(noSubMovieList))
		for index, info := range noSubMovieList {
			em.log.Debugln(index, info.VideoFileName)
		}
		em.log.Debugln("-----------------")
		em.log.Debugln("filterNoChineseSubVideoList found no chinese series", len(noSubSeriesList))
		for index, info := range noSubSeriesList {
			em.log.Debugln(index, info.VideoFileName)
		}
		em.log.Debugln("-----------------")

		// 需要将连续剧零散的每一集，进行合并到一个连续剧下面，也就是这个连续剧有那些需要更新的
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

}

// GetRecentlyAddVideoList 获取最近新添加的视频
func (em *EmbyHelper) GetRecentlyAddVideoList() ([]emby.EmbyMixInfo, []emby.EmbyMixInfo, error) {
	// 获取最近的影片列表
	items, err := em.EmbyApi.GetRecentlyItems()
	if err != nil {
		return nil, nil, err
	}
	// 获取电影和连续剧的文件夹名称
	var EpisodeIdList = make([]string, 0)
	var MovieIdList = make([]string, 0)

	em.log.Debugln("-----------------")
	em.log.Debugln("GetRecentlyAddVideoListWithNoChineseSubtitle - GetRecentlyItems Count", len(items.Items))

	// 分类
	for index, item := range items.Items {
		if item.Type == videoTypeEpisode {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			EpisodeIdList = append(EpisodeIdList, item.Id)
			em.log.Debugln("Episode:", index, item.SeriesName, item.ParentIndexNumber, item.IndexNumber)
		} else if item.Type == videoTypeMovie {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			MovieIdList = append(MovieIdList, item.Id)
			em.log.Debugln("Movie:", index, item.Name)
		} else {
			em.log.Debugln("GetRecentlyItems - Is not a goal video type:", index, item.Name, item.Type)
		}
	}

	// 过滤出有效的电影、连续剧的资源出来
	filterMovieList, err := em.getMoreVideoInfoList(MovieIdList, true)
	if err != nil {
		return nil, nil, err
	}
	filterSeriesList, err := em.getMoreVideoInfoList(EpisodeIdList, false)
	if err != nil {
		return nil, nil, err
	}
	// 输出调试信息
	em.log.Debugln("-----------------")
	em.log.Debugln("getMoreVideoInfoList found valid movie", len(filterMovieList))
	for index, info := range filterMovieList {
		em.log.Debugln(index, info.VideoFileName)
	}
	em.log.Debugln("-----------------")
	em.log.Debugln("getMoreVideoInfoList found valid series", len(filterSeriesList))
	for index, info := range filterSeriesList {
		em.log.Debugln(index, info.VideoFileName)
	}
	em.log.Debugln("-----------------")
	return filterMovieList, filterSeriesList, nil
}

// GetVideoIDPlayedMap 获取已经播放过的视频的ID
func (em *EmbyHelper) GetVideoIDPlayedMap() map[string]bool {

	videoIDPlayedMap := make(map[string]bool)
	// 获取有那些用户
	var userIds emby.EmbyUsers
	userIds, err := em.EmbyApi.GetUserIdList()
	if err != nil {
		em.log.Errorln("IsVideoIDPlayed - GetUserIdList error:", err)
		return videoIDPlayedMap
	}
	// 所有用户观看过的视频有那些
	for _, item := range userIds.Items {
		tmpRecItems, err := em.EmbyApi.GetRecentItemsByUserID(item.Id)
		if err != nil {
			em.log.Errorln("IsVideoIDPlayed - GetRecentItemsByUserID, UserID:", item.Id, "error:", err)
			return videoIDPlayedMap
		}
		// 相同的视频项目，需要判断是否已经看过了
		// 项目是否相同可以通过 Id 判断
		for _, recentlyItem := range tmpRecItems.Items {
			if recentlyItem.UserData.Played == true {
				videoIDPlayedMap[recentlyItem.Id] = true
			}
		}
	}

	return videoIDPlayedMap
}

// GetPlayedItemsSubtitle 所有用户标记播放过的视频，返回 电影、连续剧, 视频全路径 -- 对应字幕全路径（经过转换的）
func (em *EmbyHelper) GetPlayedItemsSubtitle() (map[string]string, map[string]string, error) {

	// 这个用户看过那些视频
	var userPlayedItemsList = make([]emby.UserPlayedItems, 0)
	// 获取有那些用户
	var userIds emby.EmbyUsers
	userIds, err := em.EmbyApi.GetUserIdList()
	if err != nil {
		return nil, nil, err
	}
	// 所有用户观看过的视频有那些，需要分用户统计出来
	for _, item := range userIds.Items {
		tmpRecItems, err := em.EmbyApi.GetRecentItemsByUserID(item.Id)
		if err != nil {
			return nil, nil, err
		}
		// 相同的视频项目，需要判断是否已经看过了，看过的需要排除
		// 项目是否相同可以通过 Id 判断
		oneUserPlayedItems := emby.UserPlayedItems{
			UserName: item.Name,
			UserID:   item.Id,
			Items:    make([]emby.EmbyRecentlyItem, 0),
		}
		for _, recentlyItem := range tmpRecItems.Items {

			if recentlyItem.UserData.Played == true {
				oneUserPlayedItems.Items = append(oneUserPlayedItems.Items, recentlyItem)
			}
		}

		userPlayedItemsList = append(userPlayedItemsList, oneUserPlayedItems)
	}

	var EpisodeIdList = make([]string, 0)
	var MovieIdList = make([]string, 0)
	// 把这些用户看过的视频根据 userID 和 videoID 进行查询，使用的是第几个字幕
	// 这里需要区分是 Movie 还是 Series，这样后续的路径映射才能够生效
	// 视频 emby 路径 - 字幕 emby 路径
	movieEmbyFPathMap := make(map[string]string)
	seriesEmbyFPathMap := make(map[string]string)
	for _, playedItems := range userPlayedItemsList {

		for _, item := range playedItems.Items {

			videoInfoByUserId, err := em.EmbyApi.GetItemVideoInfoByUserId(playedItems.UserID, item.Id)
			if err != nil {
				return nil, nil, err
			}

			videoInfo, err := em.EmbyApi.GetItemVideoInfo(item.Id)
			if err != nil {
				return nil, nil, err
			}
			// 首先不能越界
			if videoInfoByUserId.GetDefaultSubIndex() < 0 || len(videoInfo.MediaStreams)-1 < videoInfoByUserId.GetDefaultSubIndex() {
				em.log.Debugln("GetPlayedItemsSubtitle", videoInfo.Name, "SubIndex Out Of Range")
				continue
			}
			// 然后找出来的字幕必须是外置字幕，内置还导出个啥子
			if videoInfo.MediaStreams[videoInfoByUserId.GetDefaultSubIndex()].IsExternal == false {
				em.log.Debugln("GetPlayedItemsSubtitle", videoInfo.Name,
					"Get Played SubIndex", videoInfoByUserId.GetDefaultSubIndex(),
					"is IsExternal == false, Skip")
				continue
			}
			// 将这个字幕的 Emby 内部路径保存下来，后续还需要进行一次路径转换才能使用，转换到本程序的路径上
			if item.Type == videoTypeEpisode {
				EpisodeIdList = append(EpisodeIdList, item.Id)
				seriesEmbyFPathMap[videoInfo.Path] = videoInfo.MediaStreams[videoInfoByUserId.GetDefaultSubIndex()].Path
			} else if item.Type == videoTypeMovie {
				MovieIdList = append(MovieIdList, item.Id)
				movieEmbyFPathMap[videoInfo.Path] = videoInfo.MediaStreams[videoInfoByUserId.GetDefaultSubIndex()].Path
			}
		}
	}
	// 路径转换
	phyMovieList, err := em.getMoreVideoInfoList(MovieIdList, true)
	if err != nil {
		return nil, nil, err
	}
	phySeriesList, err := em.getMoreVideoInfoList(EpisodeIdList, false)
	if err != nil {
		return nil, nil, err
	}

	// 转换 Emby 内部路径到本程序识别的视频目录上
	// phyVideoFPath -- phySubFPath
	moviePhyFPathMap := make(map[string]string)
	seriesPhyFPathMap := make(map[string]string)
	// movie
	for _, mixInfo := range phyMovieList {
		movieEmbyFPath := mixInfo.VideoInfo.Path
		// 得到字幕的 emby 内部绝对路径
		subEmbyFPath, bok := movieEmbyFPathMap[movieEmbyFPath]
		if bok == false {
			// 没找到
			continue
		}
		movieDirFPath := filepath.Dir(mixInfo.PhysicalVideoFileFullPath)
		subFileName := filepath.Base(subEmbyFPath)
		nowPhySubFPath := filepath.Join(movieDirFPath, subFileName)
		moviePhyFPathMap[mixInfo.PhysicalVideoFileFullPath] = nowPhySubFPath
	}
	// series
	for _, mixInfo := range phySeriesList {
		epsEmbyFPath := mixInfo.VideoInfo.Path
		// 得到字幕的 emby 内部绝对路径
		subEmbyFPath, bok := seriesEmbyFPathMap[epsEmbyFPath]
		if bok == false {
			// 没找到
			continue
		}
		epsDirFPath := filepath.Dir(mixInfo.PhysicalVideoFileFullPath)
		subFileName := filepath.Base(subEmbyFPath)
		nowPhySubFPath := filepath.Join(epsDirFPath, subFileName)
		seriesPhyFPathMap[mixInfo.PhysicalVideoFileFullPath] = nowPhySubFPath
	}

	return moviePhyFPathMap, seriesPhyFPathMap, nil
}

// RefreshEmbySubList 字幕下载完毕一次，就可以触发一次这个。并发 6 线程去刷新
func (em *EmbyHelper) RefreshEmbySubList() (bool, error) {
	if em.EmbyApi == nil {
		return false, nil
	}
	err := em.EmbyApi.RefreshRecentlyVideoInfo()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (em *EmbyHelper) IsVideoPlayed(videoID string) (bool, error) {

	// 获取有那些用户
	var userIds emby.EmbyUsers
	userIds, err := em.EmbyApi.GetUserIdList()
	if err != nil {
		return false, err
	}
	// 所有用户观看过的视频有那些，需要分用户统计出来
	for _, item := range userIds.Items {
		videoInfo, err := em.EmbyApi.GetItemVideoInfoByUserId(item.Id, videoID)
		if err != nil {
			return false, err
		}
		if videoInfo.UserData.Played == true {
			return true, nil
		}
	}

	return false, nil
}

// findMappingPath 从 Emby 内置路径匹配到物理路径，返回，需要替换的前缀，以及替换到的前缀
// X:\电影    - /mnt/share1/电影
// X:\连续剧  - /mnt/share1/连续剧
func (em *EmbyHelper) findMappingPath(fileFPathWithEmby string, isMovieOrSeries bool) (bool, string, string) {

	// 这里进行路径匹配的时候需要考虑嵌套路径的问题
	// 比如，映射了 /电影  以及 /电影/AA ，那么如果有一部电影 /电影/AA/xx/xx.mkv 那么，应该匹配的是最长的路径 /电影/AA
	matchedEmbyPaths := make([]string, 0)
	if isMovieOrSeries == true {
		// 电影的情况
		for _, embyPath := range em.settings.EmbySettings.MoviePathsMapping {
			if strings.HasPrefix(fileFPathWithEmby, embyPath) == true {
				matchedEmbyPaths = append(matchedEmbyPaths, embyPath)
			}
		}
	} else {
		// 连续剧的情况
		for _, embyPath := range em.settings.EmbySettings.SeriesPathsMapping {
			if strings.HasPrefix(fileFPathWithEmby, embyPath) == true {
				matchedEmbyPaths = append(matchedEmbyPaths, embyPath)
			}
		}
	}
	if len(matchedEmbyPaths) < 1 {
		return false, "", ""
	}

	// 排序得到匹配上的路径，最长的那个
	pathSlices := sort_things.SortStringSliceByLength(matchedEmbyPaths)
	// 然后还需要从这个最长的路径，从 map 中找到对应的物理路径
	// nowPhRootPath 这个路径是映射的根目录，如果里面再次嵌套 子文件夹 再到连续剧目录，则是个问题，会丢失子文件夹目录
	nowPhRootPath := ""
	if isMovieOrSeries == true {
		// 电影的情况
		for physicalPath, embyPath := range em.settings.EmbySettings.MoviePathsMapping {
			if embyPath == pathSlices[0].Path {
				nowPhRootPath = physicalPath
				break
			}
		}
	} else {
		// 连续剧的情况
		for physicalPath, embyPath := range em.settings.EmbySettings.SeriesPathsMapping {
			if embyPath == pathSlices[0].Path {
				nowPhRootPath = physicalPath
				break
			}
		}
	}
	// 如果匹配不上
	if nowPhRootPath == "" {
		return false, "", ""
	}

	return true, pathSlices[0].Path, nowPhRootPath
}

// getVideoIMDBId 从视频的内部 ID 找到 IMDB id
func (em *EmbyHelper) getMoreVideoInfo(videoID string, isMovieOrSeries bool) (*emby.EmbyMixInfo, error) {

	if isMovieOrSeries == true {
		// 电影的情况
		info, err := em.EmbyApi.GetItemVideoInfo(videoID)
		if err != nil {
			return nil, err
		}

		ancs, err := em.EmbyApi.GetItemAncestors(videoID)
		if err != nil {
			return nil, err
		}

		mixInfo := emby.EmbyMixInfo{IMDBId: info.ProviderIds.Imdb, Ancestors: ancs, VideoInfo: info}

		return &mixInfo, nil
	} else {
		// 连续剧的情况，需要从一集对算到 series 目录，得到内部 series 的 ID，然后再得到 IMDB ID
		ancs, err := em.EmbyApi.GetItemAncestors(videoID)
		if err != nil {
			return nil, err
		}
		// 暂时不支持蓝光，因为没有下载到对应的连续剧蓝光视频
		ancestorIndex := -1
		// 找到连续剧文件夹这一层
		for i, ancestor := range ancs {
			if ancestor.Type == "Series" {
				ancestorIndex = i
				break
			}
		}
		if ancestorIndex == -1 {
			// 说明没有找到连续剧文件夹的名称，那么就应该跳过
			return nil, nil
		}
		// 这里的目标是从 Emby 获取 IMDB ID
		info, err := em.EmbyApi.GetItemVideoInfo(ancs[ancestorIndex].ID)
		if err != nil {
			return nil, err
		}
		nowSeriesIMDBID := info.ProviderIds.Imdb
		// 然后还是要跟电影一样的使用 Video ID 去获取 Ancestors 和 VideoInfo，而上面一步获取的是这个 Series 的 ID
		info, err = em.EmbyApi.GetItemVideoInfo(videoID)
		if err != nil {
			return nil, err
		}
		ancs, err = em.EmbyApi.GetItemAncestors(videoID)
		if err != nil {
			return nil, err
		}

		mixInfo := emby.EmbyMixInfo{IMDBId: nowSeriesIMDBID, Ancestors: ancs, VideoInfo: info}

		return &mixInfo, nil
	}
}

// 根据 IMDB ID 自动转换路径
func (em *EmbyHelper) autoFindMappingPathWithMixInfoByIMDBId(mixInfo *emby.EmbyMixInfo, isMovieOrSeries bool) bool {

	if mixInfo.IMDBId == "" {
		em.log.Debugln("autoFindMappingPathWithMixInfoByIMDBId", " mixInfo.IMDBId == \"\"")
		return false
	}

	// 获取 IMDB 信息
	localIMDBInfo, err := imdb_helper.GetVideoIMDBInfoFromLocal(em.log, types.VideoIMDBInfo{ImdbId: mixInfo.IMDBId}, true)
	if err != nil {

		if errors.Is(err, common2.SkipCreateInDB) == true {
			em.log.Debugln("autoFindMappingPathWithMixInfoByIMDBId.GetVideoIMDBInfoFromLocal", err)
			return false
		}

		em.log.Errorln("autoFindMappingPathWithMixInfoByIMDBId.GetVideoIMDBInfoFromLocal", err)
		return false
	}

	if localIMDBInfo.RootDirPath == "" {
		// 说明这个不是从本程序挂在的视频目录中正常扫描出来的视频，这里可能是因为 Emby 新建出来的
		em.log.Debugln("autoFindMappingPathWithMixInfoByIMDBId", " localIMDBInfo.RootDirPath == \"\"")
		return false
	}

	// 下面开始实际的路径替换，从 emby 的内部路径转换为 本程序读取到视频的路径
	if isMovieOrSeries == true {
		// 电影
		// 这里需要考虑蓝光的情况，这种目录比较特殊，在 emby 获取的时候，可以知道这个是否是蓝光，是的话，需要特殊处理
		// 伪造一个虚假不存在的 .mp4 文件向后提交给电影的下载函数
		/*
			举例：失控玩家(2021) 是一个蓝光电影
			那么下面的 mixInfo.VideoInfo.Path 从 emby 拿到应该是 /mnt/share1/电影/失控玩家(2021)
			就需要再次基础上进行视频的伪造
		*/

		if len(mixInfo.VideoInfo.MediaSources) > 0 && mixInfo.VideoInfo.MediaSources[0].Container == "bluray" {
			// 这个就是蓝光了
			// 先替换再拼接，不然会出现拼接完成后，在 Windows 下会把 /mnt/share1/电影 变为这样了 \mnt\share1\电影\失控玩家 (2021)\失控玩家 (2021).mp4
			videoReplacedDirFPath := strings.ReplaceAll(mixInfo.VideoInfo.Path, mixInfo.VideoInfo.Path, localIMDBInfo.RootDirPath)
			fakeVideoFPath := filepath.Join(videoReplacedDirFPath, filepath.Base(mixInfo.VideoInfo.Path)+common2.VideoExtMp4)
			mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(fakeVideoFPath, filepath.Dir(mixInfo.VideoInfo.Path), localIMDBInfo.RootDirPath)
			// 这个电影的文件夹
			mixInfo.VideoFolderName = filepath.Base(mixInfo.VideoInfo.Path)
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path) + common2.VideoExtMp4
		} else {
			// 常规的电影情况，也就是有一个具体的视频文件 .mp4 or .mkv
			// 在 Windows 下，如果操作 linux 的路径 /mnt/abc/123 filepath.Dir 那么得到的是 \mnt\abc
			// 那么临时方案是使用替换掉方法去得到 Dir 路径，比如是 /mnt/abc/123.mp4 replace 123.mp4 那么得到 /mnt/abc/ 还需要把最后一个字符删除
			videoDirPath := strings.ReplaceAll(mixInfo.VideoInfo.Path, filepath.Base(mixInfo.VideoInfo.Path), "")
			videoDirPath = videoDirPath[:len(videoDirPath)-1]
			mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(mixInfo.VideoInfo.Path, videoDirPath, localIMDBInfo.RootDirPath)
			// 这个电影的文件夹
			mixInfo.VideoFolderName = filepath.Base(filepath.Dir(mixInfo.VideoInfo.Path))
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
		}
	} else {
		// 连续剧
		// 暂时不支持蓝光，因为没有下载到对应的连续剧蓝光视频
		ancestorIndex := -1
		// 找到连续剧文件夹这一层
		for i, ancestor := range mixInfo.Ancestors {
			if ancestor.Type == "Series" {
				ancestorIndex = i
				break
			}
		}
		if ancestorIndex == -1 {
			// 说明没有找到连续剧文件夹的名称，那么就应该跳过
			return false
		}

		mixInfo.PhysicalSeriesRootDir = localIMDBInfo.RootDirPath
		mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(mixInfo.VideoInfo.Path, mixInfo.Ancestors[ancestorIndex].Path, localIMDBInfo.RootDirPath)
		mixInfo.PhysicalRootPath = filepath.Dir(localIMDBInfo.RootDirPath)
		// 这个剧集的文件夹
		mixInfo.VideoFolderName = filepath.Base(mixInfo.Ancestors[ancestorIndex].Path)
		mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
	}

	return true
}

// findMappingPathWithMixInfo 从 Emby 内置路径匹配到物理路径
// X:\电影    - /mnt/share1/电影
// X:\连续剧  - /mnt/share1/连续剧
func (em *EmbyHelper) findMappingPathWithMixInfo(mixInfo *emby.EmbyMixInfo, isMovieOrSeries bool) bool {

	defer func() {
		// 见 https://github.com/allanpk716/ChineseSubFinder/issues/278
		// 进行字符串操作的时候，可能会把 smb://123 转义为 smb:/123
		// 修复了个寂寞，没用，这个逻辑保留，但是没用哈。因为就不支持 SMB 的客户端协议
		if mixInfo != nil {
			mixInfo.PhysicalRootPath = path_helper.FixShareFileProtocolsPath(mixInfo.PhysicalRootPath)
			mixInfo.PhysicalVideoFileFullPath = path_helper.FixShareFileProtocolsPath(mixInfo.PhysicalVideoFileFullPath)
		}
	}()

	// 这里进行路径匹配的时候需要考虑嵌套路径的问题
	// 比如，映射了 /电影  以及 /电影/AA ，那么如果有一部电影 /电影/AA/xx/xx.mkv 那么，应该匹配的是最长的路径 /电影/AA
	matchedEmbyPaths := make([]string, 0)
	if isMovieOrSeries == true {
		// 电影的情况
		for _, embyPath := range em.settings.EmbySettings.MoviePathsMapping {
			if strings.HasPrefix(mixInfo.VideoInfo.Path, embyPath) == true {
				matchedEmbyPaths = append(matchedEmbyPaths, embyPath)
			}
		}
	} else {
		// 连续剧的情况
		for _, embyPath := range em.settings.EmbySettings.SeriesPathsMapping {
			if strings.HasPrefix(mixInfo.VideoInfo.Path, embyPath) == true {
				matchedEmbyPaths = append(matchedEmbyPaths, embyPath)
			}
		}
	}
	if len(matchedEmbyPaths) < 1 {
		return false
	}
	// 排序得到匹配上的路径，最长的那个
	pathSlices := sort_things.SortStringSliceByLength(matchedEmbyPaths)
	// 然后还需要从这个最长的路径，从 map 中找到对应的物理路径
	// nowPhRootPath 这个路径是映射的根目录，如果里面再次嵌套 子文件夹 再到连续剧目录，则是个问题，会丢失子文件夹目录
	nowPhRootPath := ""
	if isMovieOrSeries == true {
		// 电影的情况
		for physicalPath, embyPath := range em.settings.EmbySettings.MoviePathsMapping {
			if embyPath == pathSlices[0].Path {
				nowPhRootPath = physicalPath
				break
			}
		}
	} else {
		// 连续剧的情况
		for physicalPath, embyPath := range em.settings.EmbySettings.SeriesPathsMapping {
			if embyPath == pathSlices[0].Path {
				nowPhRootPath = physicalPath
				break
			}
		}
	}
	// 如果匹配不上
	if nowPhRootPath == "" {
		return false
	}

	// 下面开始实际的路径替换，从 emby 的内部路径转换为 本程序读取到视频的路径
	if isMovieOrSeries == true {
		// 电影

		// 这里需要考虑蓝光的情况，这种目录比较特殊，在 emby 获取的时候，可以知道这个是否是蓝光，是的话，需要特殊处理
		// 伪造一个虚假不存在的 .mp4 文件向后提交给电影的下载函数
		/*
			举例：失控玩家(2021) 是一个蓝光电影
			那么下面的 mixInfo.VideoInfo.Path 从 emby 拿到应该是 /mnt/share1/电影/失控玩家(2021)
			就需要再次基础上进行视频的伪造
		*/

		if len(mixInfo.VideoInfo.MediaSources) > 0 && mixInfo.VideoInfo.MediaSources[0].Container == "bluray" {
			// 这个就是蓝光了
			// 先替换再拼接，不然会出现拼接完成后，在 Windows 下会把 /mnt/share1/电影 变为这样了 \mnt\share1\电影\失控玩家 (2021)\失控玩家 (2021).mp4
			videoReplacedDirFPath := strings.ReplaceAll(mixInfo.VideoInfo.Path, pathSlices[0].Path, nowPhRootPath)
			fakeVideoFPath := filepath.Join(videoReplacedDirFPath, filepath.Base(mixInfo.VideoInfo.Path)+common2.VideoExtMp4)
			mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(fakeVideoFPath, pathSlices[0].Path, nowPhRootPath)
			// 这个电影的文件夹
			mixInfo.VideoFolderName = filepath.Base(mixInfo.VideoInfo.Path)
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path) + common2.VideoExtMp4
		} else {
			// 常规的电影情况，也就是有一个具体的视频文件 .mp4 or .mkv
			mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(mixInfo.VideoInfo.Path, pathSlices[0].Path, nowPhRootPath)
			// 因为电影搜索的时候使用的是完整的视频目录，所以这个字段并不重要，连续剧的时候才需要关注
			//mixInfo.PhysicalRootPath = strings.ReplaceAll(mixInfo.VideoInfo.Path, pathSlices[0].Path, nowPhRootPath)
			// 这个电影的文件夹
			mixInfo.VideoFolderName = filepath.Base(filepath.Dir(mixInfo.VideoInfo.Path))
			mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
		}
	} else {
		// 连续剧
		// 暂时不支持蓝光，因为没有下载到对应的连续剧蓝光视频
		ancestorIndex := -1
		// 找到连续剧文件夹这一层
		for i, ancestor := range mixInfo.Ancestors {
			if ancestor.Type == "Series" {
				ancestorIndex = i
				break
			}
		}
		if ancestorIndex == -1 {
			// 说明没有找到连续剧文件夹的名称，那么就应该跳过
			return false
		}
		mixInfo.PhysicalSeriesRootDir = strings.ReplaceAll(mixInfo.Ancestors[ancestorIndex].Path, pathSlices[0].Path, nowPhRootPath)
		mixInfo.PhysicalVideoFileFullPath = strings.ReplaceAll(mixInfo.VideoInfo.Path, pathSlices[0].Path, nowPhRootPath)
		mixInfo.PhysicalRootPath = strings.ReplaceAll(mixInfo.Ancestors[ancestorIndex+1].Path, pathSlices[0].Path, nowPhRootPath)
		// 这个剧集的文件夹
		mixInfo.VideoFolderName = filepath.Base(mixInfo.Ancestors[ancestorIndex].Path)
		mixInfo.VideoFileName = filepath.Base(mixInfo.VideoInfo.Path)
	}

	return true
}

// getMoreVideoInfoList 把视频的更多信息查询出来，需要并发去做
func (em *EmbyHelper) getMoreVideoInfoList(videoIdList []string, isMovieOrSeries bool) ([]emby.EmbyMixInfo, error) {
	var filterVideoEmbyInfo = make([]emby.EmbyMixInfo, 0)

	// 这个方法是使用两边的路径映射表来实现的转换，使用的体验不佳，很多人搞不定
	queryFuncByMatchPath := func(m string) (*emby.EmbyMixInfo, error) {

		oneMixInfo, err := em.getMoreVideoInfo(m, isMovieOrSeries)
		if err != nil {
			return nil, err
		}

		if em.settings.EmbySettings.AutoOrManual == true {
			// 通过 IMDB ID 自动转换路径
			if isMovieOrSeries == true {
				// 电影
				// 过滤掉不符合要求的,拼接绝对路径
				isFit := em.autoFindMappingPathWithMixInfoByIMDBId(oneMixInfo, isMovieOrSeries)
				if isFit == false {
					return nil, err
				}
			} else {
				// 连续剧
				// 过滤掉不符合要求的,拼接绝对路径
				isFit := em.autoFindMappingPathWithMixInfoByIMDBId(oneMixInfo, isMovieOrSeries)
				if isFit == false {
					return nil, err
				}
			}
		} else {
			// 通过手动的路径映射
			if isMovieOrSeries == true {
				// 电影
				// 过滤掉不符合要求的,拼接绝对路径
				isFit := em.findMappingPathWithMixInfo(oneMixInfo, isMovieOrSeries)
				if isFit == false {
					return nil, err
				}
			} else {
				// 连续剧
				// 过滤掉不符合要求的,拼接绝对路径
				isFit := em.findMappingPathWithMixInfo(oneMixInfo, isMovieOrSeries)
				if isFit == false {
					return nil, err
				}
			}
		}

		return oneMixInfo, nil
	}

	// em.threads
	p, err := ants.NewPoolWithFunc(em.settings.EmbySettings.Threads, func(inData interface{}) {
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
				close(done)
				close(panicChan)
			}()

			info, err := queryFuncByMatchPath(data.Id)
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
				em.log.Errorln("getMoreVideoInfoList.NewPoolWithFunc got Err", outData.Err)
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
			em.log.Errorln("getMoreVideoInfoList.NewPoolWithFunc got panic", p)
		case <-ctx.Done():
			em.log.Errorln("getMoreVideoInfoList.NewPoolWithFunc got time out", ctx.Err())
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
			em.log.Errorln("getMoreVideoInfoList ants.Invoke", err)
		}
	}
	wg.Wait()

	return filterVideoEmbyInfo, nil
}

// filterNoChineseSubVideoList 将没有中文字幕的视频找出来
func (em *EmbyHelper) filterNoChineseSubVideoList(videoList []emby.EmbyMixInfo) ([]emby.EmbyMixInfo, error) {
	currentTime := time.Now()

	var noSubVideoList = make([]emby.EmbyMixInfo, 0)
	// TODO 这里有一种情况需要考虑的，如果内置有中文的字幕，那么是否需要跳过，目前暂定的一定要有外置的字幕
	for _, info := range videoList {

		needDlSub3Month := false
		// 3个月内，或者没有字幕都要进行下载
		if info.VideoInfo.PremiereDate.AddDate(0, 0, em.settings.AdvancedSettings.TaskQueue.ExpirationTime).After(currentTime) == true {
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
				if sub_parser_hub.IsEmbySubCodecWanted(stream.Codec) == true &&
					sub_parser_hub.IsEmbySubChineseLangStringWanted(stream.Language) == true {

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
			if stream.IsExternal == false &&
				sub_parser_hub.IsEmbySubChineseLangStringWanted(stream.Language) {

				haveInsideChineseSub = true
				break
			}
		}
		// 比如，创建的时间在3个月内，然后没有额外下载的中文字幕，都符合要求
		if haveExternalChineseSub == false {
			// 没有外置字幕
			// 如果创建了7天，且有内置的中文字幕，那么也不进行下载了
			if info.VideoInfo.DateCreated.AddDate(0, 0, em.settings.AdvancedSettings.TaskQueue.DownloadSubDuringXDays).After(currentTime) == false && haveInsideChineseSub == true {
				em.log.Debugln("Create Over 7 Days, And It Has Inside ChineseSub, Than Skip", info.VideoFileName)
				continue
			}
			//// 如果创建了三个月，还是没有字幕，那么也不进行下载了
			//if info.VideoInfo.DateCreated.Add(dayRange3Months).After(currentTime) == false {
			//	continue
			//}
			// 没有中文字幕就加入下载列表
			noSubVideoList = append(noSubVideoList, info)
		} else {
			// 有外置字幕
			// 如果视频发布时间超过两年了，有字幕就直接跳过了，一般字幕稳定了
			if currentTime.Year()-2 > info.VideoInfo.PremiereDate.Year() {
				em.log.Debugln("Create Over 2 Years, And It Has External ChineseSub, Than Skip", info.VideoFileName)
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

// GetInternalEngSubAndExChineseEnglishSub 获取对应 videoId 的内置英文字幕，外置中文字幕（只要是带有中文的都算，简体、繁体、简英、繁英，需要后续额外的判断）字幕
func (em *EmbyHelper) GetInternalEngSubAndExChineseEnglishSub(videoId string) (bool, []emby.SubInfo, []emby.SubInfo, error) {

	// 先刷新以下这个资源，避免找到的字幕不存在了
	err := em.EmbyApi.UpdateVideoSubList(videoId)
	if err != nil {
		return false, nil, nil, err
	}
	// 获取这个资源的信息
	videoInfo, err := em.EmbyApi.GetItemVideoInfo(videoId)
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
		if stream.IsExternal == false && stream.Language == language.Emby_English_eng && stream.Codec == streamCodec {
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
			if sub_parser_hub.IsEmbySubCodecWanted(stream.Codec) == true &&
				sub_parser_hub.IsEmbySubChineseLangStringWanted(stream.Language) == true {

				tmpFileName := filepath.Base(stream.Path)
				// 去除 .default 或者 .forced
				//tmpFileName = strings.ReplaceAll(tmpFileName, subparser.Sub_Ext_Mark_Default, "")
				//tmpFileName = strings.ReplaceAll(tmpFileName, subparser.Sub_Ext_Mark_Forced, "")
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
			subFileData, err := em.EmbyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", index), common2.SubExtSRT)
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
		tmpExt := common2.SubExtSRT
		if i == 1 {
			tmpExt = common2.SubExtASS
		}
		subFileData, err := em.EmbyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", InsideEngSubIndex), tmpExt)
		if err != nil {
			return false, nil, nil, err
		}
		tmpInSubInfo := emby.NewSubInfo(videoFileNameWithOutExt+tmpExt, tmpExt, InsideEngSubIndex)
		tmpInSubInfo.Content = []byte(subFileData)
		inSubList = append(inSubList, *tmpInSubInfo)
	}
	// 再下载外置的
	for i, subInfo := range exSubList {
		subFileData, err := em.EmbyApi.GetSubFileData(videoId, mediaSourcesId, fmt.Sprintf("%d", subInfo.EmbyStreamIndex), subInfo.Ext)
		if err != nil {
			return false, nil, nil, err
		}
		exSubList[i].Content = []byte(subFileData)
	}

	return true, inSubList, exSubList, nil
}

// CheckPath 检查路径 EmbyConfig 配置中的映射路径是否是有效的，
func (em *EmbyHelper) CheckPath(pathType string) ([]string, error) {

	// 获取最近的影片列表
	items, err := em.EmbyApi.GetRecentlyItems()
	if err != nil {
		return nil, err
	}
	// 获取电影和连续剧的文件夹名称
	var EpisodeIdList = make([]string, 0)
	var MovieIdList = make([]string, 0)
	// 分类
	for index, item := range items.Items {
		if item.Type == videoTypeEpisode {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			EpisodeIdList = append(EpisodeIdList, item.Id)
			em.log.Debugln("Episode:", index, item.SeriesName, item.ParentIndexNumber, item.IndexNumber)
		} else if item.Type == videoTypeMovie {
			// 这个里面可能混有其他的内容，比如目标是连续剧，但是 emby_helper 其实会把其他的混合内容也标记进去
			MovieIdList = append(MovieIdList, item.Id)
			em.log.Debugln("Movie:", index, item.Name)
		} else {
			em.log.Debugln("GetRecentlyItems - Is not a goal video type:", index, item.Name, item.Type)
		}
	}

	outCount := 0
	outList := make([]string, 0)

	if pathType == "movie" {
		// 过滤出有效的电影、连续剧的资源出来
		filterMovieList, err := em.getMoreVideoInfoList(MovieIdList, true)
		if err != nil {
			return nil, err
		}

		for _, info := range filterMovieList {

			if my_util.IsFile(info.PhysicalVideoFileFullPath) == true {
				outList = append(outList, info.PhysicalVideoFileFullPath)
				outCount++
				if outCount > 5 {
					break
				}
			}
		}

	} else {
		filterSeriesList, err := em.getMoreVideoInfoList(EpisodeIdList, false)
		if err != nil {
			return nil, err
		}

		for _, info := range filterSeriesList {

			if my_util.IsFile(info.PhysicalVideoFileFullPath) == true {
				outList = append(outList, info.PhysicalVideoFileFullPath)
				outCount++
				if outCount > 5 {
					break
				}
			}
		}
	}

	return outList, nil
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
	streamCodec      = "subrip"
)
