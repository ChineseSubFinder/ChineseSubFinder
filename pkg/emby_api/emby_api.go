package emby_api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/emby"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type EmbyApi struct {
	log     *logrus.Logger
	timeOut time.Duration
}

func NewEmbyApi(log *logrus.Logger) *EmbyApi {
	em := EmbyApi{}
	em.log = log
	// 检查是否超过范围
	settings.Get().Check()
	// 强制设置
	em.timeOut = 5 * 60 * time.Second
	return &em
}

// RefreshRecentlyVideoInfo 字幕下载完毕一次，就可以触发一次这个。并发 6 线程去刷新
func (em *EmbyApi) RefreshRecentlyVideoInfo(embySettings *settings.EmbySettings, SkipWatched bool, maxRequestVideoNumber int) error {
	items, err := em.GetRecentlyItems(embySettings, SkipWatched, maxRequestVideoNumber)
	if err != nil {
		return err
	}

	em.log.Debugln("RefreshRecentlyVideoInfo - GetRecentlyItems Count", len(items.Items))

	updateFunc := func(i interface{}) error {
		tmpId := i.(string)
		return em.UpdateVideoSubList(embySettings, tmpId)
	}
	p, err := ants.NewPoolWithFunc(embySettings.Threads, func(inData interface{}) {
		data := inData.(InputData)
		defer data.Wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), em.timeOut)
		defer cancel()

		done := make(chan error, 1)
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}

				close(done)
				close(panicChan)
			}()

			done <- updateFunc(data.Id)
		}()

		select {
		case err = <-done:
			if err != nil {
				em.log.Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got error", err)
			}
			return
		case p := <-panicChan:
			em.log.Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got panic", p)
		case <-ctx.Done():
			em.log.Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got time out", ctx.Err())
			return
		}
	})
	if err != nil {
		return err
	}
	defer p.Release()
	wg := sync.WaitGroup{}
	for _, item := range items.Items {
		wg.Add(1)
		err = p.Invoke(InputData{Id: item.Id, Wg: &wg})
		if err != nil {
			em.log.Errorln("RefreshRecentlyVideoInfo ants.Invoke", err)
		}
	}
	wg.Wait()

	return nil
}

func (em *EmbyApi) GetRecentItemsByUserID(embySettings *settings.EmbySettings, userId string, maxRequestVideoNumber int) (emby.EmbyRecentlyItems, error) {

	var tmpRecItems emby.EmbyRecentlyItems
	// 获取指定用户的视频列表
	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"api_key":          embySettings.APIKey,
			"IsUnaired":        "false",
			"Limit":            fmt.Sprintf("%d", maxRequestVideoNumber),
			"Recursive":        "true",
			"SortOrder":        "Descending",
			"IncludeItemTypes": "Episode,Movie",
			"Filters":          "IsNotFolder",
			"SortBy":           "DateCreated",
		}).
		SetResult(&tmpRecItems).
		Get(embySettings.AddressUrl + "/emby/Users/" + userId + "/Items")
	if err != nil {
		return emby.EmbyRecentlyItems{}, err
	}

	return tmpRecItems, nil
}

// GetRecentlyItems 获取近期的视频(根据 SkipWatched 的情况，如果不跳过，那么就是获取所有用户的列表，如果是跳过，那么就会单独读取每个用户的再交叉判断)
// 在 API 调试界面 -- ItemsService
func (em *EmbyApi) GetRecentlyItems(embySettings *settings.EmbySettings, SkipWatched bool, maxRequestVideoNumber int) (emby.EmbyRecentlyItems, error) {

	var recItems emby.EmbyRecentlyItems
	recItems.Items = make([]emby.EmbyRecentlyItem, 0)
	var recItemMap = make(map[string]emby.EmbyRecentlyItem)
	var recItemExsitMap = make(map[string]emby.EmbyRecentlyItem)
	var err error
	if SkipWatched == false {
		em.log.Debugln("Emby Setting SkipWatched = false")

		// 默认是不指定某一个User的视频列表
		_, err = em.createClient().R().
			SetQueryParams(map[string]string{
				"api_key":          embySettings.APIKey,
				"IsUnaired":        "false",
				"Limit":            fmt.Sprintf("%d", maxRequestVideoNumber),
				"Recursive":        "true",
				"SortOrder":        "Descending",
				"IncludeItemTypes": "Episode,Movie",
				"Filters":          "IsNotFolder",
				"SortBy":           "DateCreated",
			}).
			SetResult(&recItems).
			Get(embySettings.AddressUrl + "/emby/Items")

		if err != nil {
			return emby.EmbyRecentlyItems{}, err
		}
	} else {
		em.log.Debugln("Emby Setting SkipWatched = true")

		var userIds emby.EmbyUsers
		userIds, err = em.GetUserIdList(embySettings)
		if err != nil {
			return emby.EmbyRecentlyItems{}, err
		}

		for _, item := range userIds.Items {

			tmpRecItems, err := em.GetRecentItemsByUserID(embySettings, item.Id, maxRequestVideoNumber)
			if err != nil {
				return emby.EmbyRecentlyItems{}, err
			}
			// 相同的视频项目，需要判断是否已经看过了，看过的需要排除
			// 项目是否相同可以通过 Id 判断
			for _, recentlyItem := range tmpRecItems.Items {
				// 这个视频是否已经插入过了，可能会进行删除
				_, bFound := recItemMap[recentlyItem.Id]
				if bFound == false {
					// map 中不存在
					// 如果没有播放过，则插入
					if recentlyItem.UserData.Played == false {
						recItemMap[recentlyItem.Id] = recentlyItem
					}
				} else {
					// map 中存在
					// 既然存在，则可以理解为其他人是没有看过的，但是，如果当前的用户看过了，那么就要删除这一条
					if recentlyItem.UserData.Played == true {
						// 先记录下来，然后再删除这一条
						recItemExsitMap[recentlyItem.Id] = recentlyItem
					}
				}
				recItemMap[recentlyItem.Id] = recentlyItem
			}
		}

		for id := range recItemExsitMap {
			em.log.Debugln("Skip Watched Video:", recItemMap[id].Type, recItemMap[id].Name)
			delete(recItemMap, id)
		}

		for _, item := range recItemMap {
			recItems.Items = append(recItems.Items, item)
		}

		recItems.TotalRecordCount = len(recItemMap)
	}

	return recItems, nil
}

// GetUserIdList 获取所有的 UserId
func (em *EmbyApi) GetUserIdList(embySettings *settings.EmbySettings) (emby.EmbyUsers, error) {
	var recItems emby.EmbyUsers

	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"api_key": embySettings.APIKey,
		}).
		SetResult(&recItems).
		Get(embySettings.AddressUrl + "/emby/Users/Query")

	if err != nil {
		return emby.EmbyUsers{}, err
	}
	return recItems, nil
}

// GetItemAncestors 获取父级信息，在 API 调试界面 -- LibraryService
func (em *EmbyApi) GetItemAncestors(embySettings *settings.EmbySettings, id string) ([]emby.EmbyItemsAncestors, error) {

	var recItems []emby.EmbyItemsAncestors

	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"api_key": embySettings.APIKey,
		}).
		SetResult(&recItems).
		Get(embySettings.AddressUrl + "/emby/Items/" + id + "/Ancestors")
	if err != nil {
		return nil, err
	}

	return recItems, nil
}

// GetItemVideoInfo 在 API 调试界面 -- UserLibraryService，如果是电影，那么是可以从 ProviderIds 得到 IMDB ID 的
// 如果是连续剧，那么不能使用一集的ID取获取，需要是这个剧集的 ID，注意一季的ID也是不行的
func (em *EmbyApi) GetItemVideoInfo(embySettings *settings.EmbySettings, id string) (emby.EmbyVideoInfo, error) {

	var recItem emby.EmbyVideoInfo

	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"api_key": embySettings.APIKey,
		}).
		SetResult(&recItem).
		Get(embySettings.AddressUrl + "/emby/LiveTv/Programs/" + id)
	if err != nil {
		return emby.EmbyVideoInfo{}, err
	}

	return recItem, nil
}

// GetItemVideoInfoByUserId 可以拿到这个视频的选择字幕Index，配合 GetItemVideoInfo 使用。 在 API 调试界面 -- UserLibraryService
func (em *EmbyApi) GetItemVideoInfoByUserId(embySettings *settings.EmbySettings, userId, videoId string) (emby.EmbyVideoInfoByUserId, error) {

	var recItem emby.EmbyVideoInfoByUserId

	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"api_key": embySettings.APIKey,
		}).
		SetResult(&recItem).
		Get(embySettings.AddressUrl + "/emby/Users/" + userId + "/Items/" + videoId)
	if err != nil {
		return emby.EmbyVideoInfoByUserId{}, err
	}

	return recItem, nil
}

// UpdateVideoSubList 更新字幕列表， 在 API 调试界面 -- ItemRefreshService
func (em *EmbyApi) UpdateVideoSubList(embySettings *settings.EmbySettings, id string) error {

	_, err := em.createClient().R().
		SetQueryParams(map[string]string{
			"Recursive": "true",
			"api_key":   embySettings.APIKey,
		}).
		Post(embySettings.AddressUrl + "/emby/Items/" + id + "/Refresh")
	if err != nil {
		return err
	}

	return nil
}

// GetSubFileData 下载字幕 subExt -> .ass or .srt , 在 API 调试界面 -- SubtitleService
func (em *EmbyApi) GetSubFileData(embySettings *settings.EmbySettings, videoId, mediaSourceId, subIndex, subExt string) (string, error) {

	response, err := em.createClient().R().
		Get(embySettings.AddressUrl + "/emby/Videos/" + videoId + "/" + mediaSourceId + "/Subtitles/" + subIndex + "/Stream" + subExt)
	if err != nil {
		return "", err
	}

	return response.String(), nil
}

func (em *EmbyApi) createClient() *resty.Client {
	// 见 https://github.com/ChineseSubFinder/ChineseSubFinder/issues/140
	client := resty.New().SetTransport(&http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}).RemoveProxy().SetTimeout(em.timeOut)
	return client
}

type InputData struct {
	Id string
	Wg *sync.WaitGroup
}
