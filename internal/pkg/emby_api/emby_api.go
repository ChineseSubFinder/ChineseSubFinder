package emby_api

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type EmbyApi struct {
	embyConfig emby.EmbyConfig
	threads    int
	timeOut    time.Duration
}

func NewEmbyApi(embyConfig emby.EmbyConfig) *EmbyApi {
	em := EmbyApi{}
	em.embyConfig = embyConfig
	if em.embyConfig.LimitCount < common.EmbyApiGetItemsLimitMin ||
		em.embyConfig.LimitCount > common.EmbyApiGetItemsLimitMax {

		em.embyConfig.LimitCount = common.EmbyApiGetItemsLimitMin
	}
	em.threads = 6
	em.timeOut = 5 * 60 * time.Second
	return &em
}

// RefreshRecentlyVideoInfo 字幕下载完毕一次，就可以触发一次这个。并发 6 线程去刷新
func (em EmbyApi) RefreshRecentlyVideoInfo() error {
	items, err := em.GetRecentlyItems()
	if err != nil {
		return err
	}

	updateFunc := func(i interface{}) error {
		tmpId := i.(string)
		return em.UpdateVideoSubList(tmpId)
	}
	p, err := ants.NewPoolWithFunc(em.threads, func(inData interface{}) {
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
			}()

			done <- updateFunc(data.Id)
		}()

		select {
		case err = <-done:
			if err != nil {
				log_helper.GetLogger().Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got error", err)
			}
			return
		case p := <-panicChan:
			log_helper.GetLogger().Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got panic", p)
		case <-ctx.Done():
			log_helper.GetLogger().Errorln("RefreshRecentlyVideoInfo.NewPoolWithFunc got time out", ctx.Err())
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
			log_helper.GetLogger().Errorln("RefreshRecentlyVideoInfo ants.Invoke", err)
		}
	}
	wg.Wait()

	return nil
}

// GetRecentlyItems 获取近期的视频，在 API 调试界面 -- ItemsService
func (em EmbyApi) GetRecentlyItems() (emby.EmbyRecentlyItems, error) {

	var recItems emby.EmbyRecentlyItems
	recItems.Items = make([]emby.EmbyRecentlyItem, 0)
	var recItemMap = make(map[string]emby.EmbyRecentlyItem)
	var recItemExsitMap = make(map[string]emby.EmbyRecentlyItem)
	var err error
	if em.embyConfig.SkipWatched == false {
		// 默认是不指定某一个User的视频列表
		_, err = em.getNewClient().R().
			SetQueryParams(map[string]string{
				"api_key":          em.embyConfig.ApiKey,
				"IsUnaired":        "false",
				"Limit":            fmt.Sprintf("%d", em.embyConfig.LimitCount),
				"Recursive":        "true",
				"SortOrder":        "Descending",
				"IncludeItemTypes": "Episode,Movie",
				"Filters":          "IsNotFolder",
				"SortBy":           "DateCreated",
			}).
			SetResult(&recItems).
			Get(em.embyConfig.Url + "/emby/Items")
	} else {

		var userIds emby.EmbyUsers
		userIds, err = em.GetUserIdList()
		if err != nil {
			return emby.EmbyRecentlyItems{}, err
		}

		for _, item := range userIds.Items {
			var tmpRecItems emby.EmbyRecentlyItems
			// 获取指定用户的视频列表
			_, err = em.getNewClient().R().
				SetQueryParams(map[string]string{
					"api_key":          em.embyConfig.ApiKey,
					"IsUnaired":        "false",
					"Limit":            fmt.Sprintf("%d", em.embyConfig.LimitCount),
					"Recursive":        "true",
					"SortOrder":        "Descending",
					"IncludeItemTypes": "Episode,Movie",
					"Filters":          "IsNotFolder",
					"SortBy":           "DateCreated",
				}).
				SetResult(&tmpRecItems).
				Get(em.embyConfig.Url + "/emby/Users/" + item.Id + "/Items")

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
func (em EmbyApi) GetUserIdList() (emby.EmbyUsers, error) {
	var recItems emby.EmbyUsers
	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		SetResult(&recItems).
		Get(em.embyConfig.Url + "/emby/Users/Query")

	if err != nil {
		return emby.EmbyUsers{}, err
	}

	return recItems, nil
}

// GetItemAncestors 获取父级信息，在 API 调试界面 -- LibraryService
func (em EmbyApi) GetItemAncestors(id string) ([]emby.EmbyItemsAncestors, error) {

	var recItems []emby.EmbyItemsAncestors

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		SetResult(&recItems).
		Get(em.embyConfig.Url + "/emby/Items/" + id + "/Ancestors")
	if err != nil {
		return nil, err
	}

	return recItems, nil
}

// GetItemVideoInfo 在 API 调试界面 -- UserLibraryService
func (em EmbyApi) GetItemVideoInfo(id string) (emby.EmbyVideoInfo, error) {

	var recItem emby.EmbyVideoInfo

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		SetResult(&recItem).
		Get(em.embyConfig.Url + "/emby/LiveTv/Programs/" + id)
	if err != nil {
		return emby.EmbyVideoInfo{}, err
	}

	return recItem, nil
}

// GetItemVideoInfoByUserId 可以拿到这个视频的选择字幕Index，配合 GetItemVideoInfo 使用。 在 API 调试界面 -- UserLibraryService
func (em EmbyApi) GetItemVideoInfoByUserId(userId, videoId string) (emby.EmbyVideoInfoByUserId, error) {

	var recItem emby.EmbyVideoInfoByUserId

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		SetResult(&recItem).
		Get(em.embyConfig.Url + "/emby/Users/" + userId + "/Items/" + videoId)
	if err != nil {
		return emby.EmbyVideoInfoByUserId{}, err
	}

	return recItem, nil
}

// UpdateVideoSubList 更新字幕列表， 在 API 调试界面 -- ItemRefreshService
func (em EmbyApi) UpdateVideoSubList(id string) error {

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		Post(em.embyConfig.Url + "/emby/Items/" + id + "/Refresh")
	if err != nil {
		return err
	}

	return nil
}

// GetSubFileData 下载字幕 subExt -> .ass or .srt , 在 API 调试界面 -- SubtitleService
func (em EmbyApi) GetSubFileData(videoId, mediaSourceId, subIndex, subExt string) (string, error) {

	response, err := em.getNewClient().R().
		Get(em.embyConfig.Url + "/emby/Videos/" + videoId + "/" + mediaSourceId + "/Subtitles/" + subIndex + "/Stream" + subExt)
	if err != nil {
		return "", err
	}

	return response.String(), nil
}

func (em EmbyApi) getNewClient() *resty.Client {
	tmpClient := resty.New()
	tmpClient.RemoveProxy()
	tmpClient.SetTimeout(em.timeOut)
	return tmpClient
}

type InputData struct {
	Id string
	Wg *sync.WaitGroup
}
