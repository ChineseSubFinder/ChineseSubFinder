package file_downloader

import (
	"crypto/sha256"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/download_file_cache"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type FileDownloader struct {
	Settings          *settings.Settings
	Log               *logrus.Logger
	downloadFileCache *download_file_cache.DownloadFileCache
}

func NewFileDownloader(settings *settings.Settings, log *logrus.Logger) *FileDownloader {
	return &FileDownloader{Settings: settings, Log: log,
		downloadFileCache: download_file_cache.NewDownloadFileCache(settings)}
}

// Get supplierName 这个参数一定得是字幕源的名称，通过 s.GetSupplierName() 获取，否则后续的字幕源今日下载量将不能正确统计和判断
// xunlei、shooter 使用这个
func (f *FileDownloader) Get(supplierName string, topN int64, videoFileName string,
	fileDownloadUrl string, score int64, offset int64) (*supplier.SubInfo, error) {

	fileUID := fmt.Sprintf("%x", sha256.Sum256([]byte(fileDownloadUrl)))

	found, subInfo, err := f.downloadFileCache.Get(fileUID)
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
		_, err = task_queue.AddDailyDownloadCount(supplierName,
			my_util.GetPublicIP(f.Settings.AdvancedSettings.TaskQueue, f.Settings.AdvancedSettings.ProxySettings))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.Get.AddDailyDownloadCount", err)
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
		err = f.downloadFileCache.Add(inSubInfo)
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
	found, subInfo, err := f.downloadFileCache.Get(fileUID)
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
		_, err = task_queue.AddDailyDownloadCount(supplierName,
			my_util.GetPublicIP(f.Settings.AdvancedSettings.TaskQueue, f.Settings.AdvancedSettings.ProxySettings))
		if err != nil {
			f.Log.Warningln(supplierName, "FileDownloader.GetEx.AddDailyDownloadCount", err)
		}
		// 默认存入都是简体中文的语言类型，后续取出来的时候需要再次调用 SubParser 进行解析
		err = f.downloadFileCache.Add(subInfo)
		if err != nil {
			return nil, err
		}

		return subInfo, nil
	} else {
		// 如果已经存在缓存中，那么就直接返回
		return subInfo, nil
	}
}
