package file_downloader

import (
	"crypto/sha256"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"path/filepath"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/subtitle_best_api"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

type FileDownloader struct {
	Settings        *settings.Settings
	Log             *logrus.Logger
	CacheCenter     *cache_center.CacheCenter
	SubParserHub    *sub_parser_hub.SubParserHub
	SubtitleBestApi *subtitle_best_api.SubtitleBestApi
}

func NewFileDownloader(cacheCenter *cache_center.CacheCenter, authKey random_auth_key.AuthKey) *FileDownloader {

	f := FileDownloader{Settings: cacheCenter.Settings,
		Log:             cacheCenter.Log,
		CacheCenter:     cacheCenter,
		SubParserHub:    sub_parser_hub.NewSubParserHub(cacheCenter.Log, ass.NewParser(cacheCenter.Log), srt.NewParser(cacheCenter.Log)),
		SubtitleBestApi: subtitle_best_api.NewSubtitleBestApi(authKey, cacheCenter.Settings.AdvancedSettings.ProxySettings),
	}
	return &f
}

func (f *FileDownloader) GetName() string {
	return f.CacheCenter.GetName()
}

// Get supplierName 这个参数一定得是字幕源的名称，通过 s.GetSupplierName() 获取，否则后续的字幕源今日下载量将不能正确统计和判断
// xunlei、shooter 使用这个
func (f *FileDownloader) Get(supplierName string, topN int64, videoFileName string,
	fileDownloadUrl string, score int64, offset int64, cacheString ...string) (*supplier.SubInfo, error) {

	var fileUID string

	if len(cacheString) < 1 {
		fileUID = fmt.Sprintf("%x", sha256.Sum256([]byte(fileDownloadUrl)))
	} else {
		fileUID = cacheString[0]
	}

	found, subInfo, err := f.CacheCenter.DownloadFileGet(fileUID)
	if err != nil {
		return nil, err
	}
	// 如果不存在那么就先下载，然后再存入缓存中
	if found == false {
		fileData, downloadFileName, err := my_util.DownFile(f.Log, fileDownloadUrl, f.Settings.AdvancedSettings.ProxySettings)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			my_util.GetPublicIP(f.Log, f.Settings.AdvancedSettings.TaskQueue, f.Settings.AdvancedSettings.ProxySettings))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.Get.DailyDownloadCountAdd", err)
		}
		// 需要获取下载文件的后缀名，后续才指导是要解压还是直接解析字幕
		ext := ""
		if downloadFileName == "" {
			ext = filepath.Ext(fileDownloadUrl)
		} else {
			ext = filepath.Ext(downloadFileName)
		}
		// 默认存入都是简体中文的语言类型，后续取出来的时候需要再次调用 SubParser 进行解析
		inSubInfo := supplier.NewSubInfo(supplierName, topN, videoFileName, language.ChineseSimple, fileDownloadUrl, score, offset, ext, fileData)

		if len(cacheString) > 0 {
			// 专门为 ASSRT 这种下载连接是临时情况而定制的
			inSubInfo.SetFileUrlSha256(fileUID)
		}

		err = f.CacheCenter.DownloadFileAdd(inSubInfo)
		if err != nil {
			return nil, err
		}

		return inSubInfo, nil
	} else {
		// 如果已经存在缓存中，那么就直接返回
		return subInfo, nil
	}
}

// GetA4k supplierName 这个参数一定得是字幕源的名称，通过 s.GetSupplierName() 获取，否则后续的字幕源今日下载量将不能正确统计和判断
func (f *FileDownloader) GetA4k(supplierName string, topN int64, season, eps int,
	videoFileName string, fileDownloadUrl string) (*supplier.SubInfo, error) {

	var fileUID string
	fileUID = fmt.Sprintf("%x", sha256.Sum256([]byte(fileDownloadUrl)))

	found, subInfo, err := f.CacheCenter.DownloadFileGet(fileUID)
	if err != nil {
		return nil, err
	}
	// 如果不存在那么就先下载，然后再存入缓存中
	if found == false {
		fileData, downloadFileName, err := my_util.DownFile(f.Log, fileDownloadUrl, f.Settings.AdvancedSettings.ProxySettings)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			my_util.GetPublicIP(f.Log, f.Settings.AdvancedSettings.TaskQueue, f.Settings.AdvancedSettings.ProxySettings))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.Get.DailyDownloadCountAdd", err)
		}
		// 需要获取下载文件的后缀名，后续才指导是要解压还是直接解析字幕
		ext := ""
		if downloadFileName == "" {
			ext = filepath.Ext(fileDownloadUrl)
		} else {
			ext = filepath.Ext(downloadFileName)
		}
		// 默认存入都是简体中文的语言类型，后续取出来的时候需要再次调用 SubParser 进行解析
		inSubInfo := supplier.NewSubInfo(supplierName, topN, videoFileName, language.ChineseSimple, fileDownloadUrl, 0, 0, ext, fileData)
		inSubInfo.Season = season
		inSubInfo.Episode = eps
		inSubInfo.GetUID()

		err = f.CacheCenter.DownloadFileAdd(inSubInfo)
		if err != nil {
			return nil, err
		}

		return inSubInfo, nil
	} else {
		// 如果已经存在缓存中，那么就直接返回
		return subInfo, nil
	}
}

// GetEx supplierName 这个参数一定得是字幕源的名称，通过 s.GetSupplierName() 获取，否则后续的字幕源今日下载量将不能正确统计和判断
// zimuku、subhd 使用这个
func (f *FileDownloader) GetEx(supplierName string, browser *rod.Browser, subDownloadPageUrl string, TopN int64, Season, Episode int, downFileFunc func(browser *rod.Browser, subDownloadPageUrl string, TopN int64, Season, Episode int) (*supplier.SubInfo, error)) (*supplier.SubInfo, error) {

	fileUID := fmt.Sprintf("%x", sha256.Sum256([]byte(subDownloadPageUrl)))
	found, subInfo, err := f.CacheCenter.DownloadFileGet(fileUID)
	if err != nil {
		return nil, err
	}
	// 如果不存在那么就先下载，然后再存入缓存中
	if found == false {

		subInfo, err = downFileFunc(browser, subDownloadPageUrl, TopN, Season, Episode)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			my_util.GetPublicIP(f.Log, f.Settings.AdvancedSettings.TaskQueue, f.Settings.AdvancedSettings.ProxySettings))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.GetEx.DailyDownloadCountAdd", err)
		}
		// 默认存入都是简体中文的语言类型，后续取出来的时候需要再次调用 SubParser 进行解析
		err = f.CacheCenter.DownloadFileAdd(subInfo)
		if err != nil {
			return nil, err
		}

		return subInfo, nil
	} else {
		// 如果已经存在缓存中，那么就直接返回
		return subInfo, nil
	}
}

// GetCSF subtitle.best 使用这个
func (f *FileDownloader) GetCSF(subSha256 string) (bool, *supplier.SubInfo, error) {

	found, subInfo, err := f.CacheCenter.DownloadFileGet(subSha256)
	if err != nil {
		return false, nil, err
	}

	if found == false {
		// 没有找到就是缓存进去
		return false, nil, nil
	} else {
		// 缓存中有
		return true, subInfo, nil
	}
}

// AddCSF subtitle.best 使用这个
func (f FileDownloader) AddCSF(inSubInfo *supplier.SubInfo) error {

	inSubInfo.SetFileUrlSha256(inSubInfo.FileUrl)
	err := f.CacheCenter.DownloadFileAdd(inSubInfo)
	if err != nil {
		return err
	}
	return nil
}
