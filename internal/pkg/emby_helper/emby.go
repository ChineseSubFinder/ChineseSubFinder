package emby_helper

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

func NewEmbyHelper(embyConfig emby.EmbyConfig) *EmbyApi {
	em := EmbyApi{}
	em.embyConfig = embyConfig
	if em.embyConfig.LimitCount < common.EmbyApiGetItemsLimitMin ||
		em.embyConfig.LimitCount > common.EmbyApiGetItemsLimitMax {

		em.embyConfig.LimitCount = common.EmbyApiGetItemsLimitMin
	}
	em.threads = 6
	em.timeOut = 5 * time.Second
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

// GetRecentlyItems 在 API 调试界面 -- ItemsService
func (em EmbyApi) GetRecentlyItems() (emby.EmbyRecentlyItems, error) {

	var recItems emby.EmbyRecentlyItems
	_, err := em.getNewClient().R().
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
		Get(em.embyConfig.Url + "/emby_helper/Items")
	if err != nil {
		return emby.EmbyRecentlyItems{}, err
	}

	return recItems, nil
}

// GetItemAncestors 在 API 调试界面 -- LibraryService
func (em EmbyApi) GetItemAncestors(id string) ([]emby.EmbyItemsAncestors, error) {

	var recItems []emby.EmbyItemsAncestors

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		SetResult(&recItems).
		Get(em.embyConfig.Url + "/emby_helper/Items/" + id + "/Ancestors")
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
		Get(em.embyConfig.Url + "/emby_helper/LiveTv/Programs/" + id)
	if err != nil {
		return emby.EmbyVideoInfo{}, err
	}

	return recItem, nil
}

// UpdateVideoSubList 在 API 调试界面 -- ItemRefreshService
func (em EmbyApi) UpdateVideoSubList(id string) error {

	_, err := em.getNewClient().R().
		SetQueryParams(map[string]string{
			"api_key": em.embyConfig.ApiKey,
		}).
		Post(em.embyConfig.Url + "/emby_helper/Items/" + id + "/Refresh")
	if err != nil {
		return err
	}

	return nil
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
