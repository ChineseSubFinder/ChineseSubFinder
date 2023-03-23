package file_downloader

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/subtitle_best_api"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

type FileDownloader struct {
	Log              *logrus.Logger
	CacheCenter      *cache_center.CacheCenter
	SubParserHub     *sub_parser_hub.SubParserHub
	MediaInfoDealers *media_info_dealers.Dealers
}

func NewFileDownloader(cacheCenter *cache_center.CacheCenter, authKey random_auth_key.AuthKey) *FileDownloader {

	f := FileDownloader{
		Log:              cacheCenter.Log,
		CacheCenter:      cacheCenter,
		SubParserHub:     sub_parser_hub.NewSubParserHub(cacheCenter.Log, ass.NewParser(cacheCenter.Log), srt.NewParser(cacheCenter.Log)),
		MediaInfoDealers: media_info_dealers.NewDealers(cacheCenter.Log, subtitle_best_api.NewSubtitleBestApi(cacheCenter.Log, authKey)),
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
		fileData, downloadFileName, err := pkg.DownFile(f.Log, fileDownloadUrl)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			pkg.GetPublicIP(f.Log, settings.Get().AdvancedSettings.TaskQueue))
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
		fileData, downloadFileName, err := pkg.DownFile(f.Log, fileDownloadUrl)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			pkg.GetPublicIP(f.Log, settings.Get().AdvancedSettings.TaskQueue))
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
			pkg.GetPublicIP(f.Log, settings.Get().AdvancedSettings.TaskQueue))
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

// GetSubtitleBest supplierName 这个参数一定得是字幕源的名称，通过 s.GetSupplierName() 获取，否则后续的字幕源今日下载量将不能正确统计和判断
func (f *FileDownloader) GetSubtitleBest(supplierName string, topN int64, season, eps int,
	title, ext, subSha256, fileDownloadUrl string) (*supplier.SubInfo, error) {

	found, subInfo, err := f.CacheCenter.DownloadFileGet(subSha256)
	if err != nil {
		return nil, err
	}
	// 如果不存在那么就先下载，然后再存入缓存中
	if found == false {

		fileData, _, err := pkg.DownFile(f.Log, fileDownloadUrl)
		if err != nil {
			return nil, err
		}
		// 下载成功需要统计到今天的次数中
		_, err = f.CacheCenter.DailyDownloadCountAdd(supplierName,
			pkg.GetPublicIP(f.Log, settings.Get().AdvancedSettings.TaskQueue))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.Get.DailyDownloadCountAdd", err)
		}
		// 默认存入都是简体中文的语言类型，后续取出来的时候需要再次调用 SubParser 进行解析
		inSubInfo := supplier.NewSubInfo(supplierName, topN, title, language.ChineseSimple, fileDownloadUrl, 0, 0, ext, fileData)
		inSubInfo.Season = season
		inSubInfo.Episode = eps
		inSubInfo.SetFileUrlSha256(subSha256)
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
