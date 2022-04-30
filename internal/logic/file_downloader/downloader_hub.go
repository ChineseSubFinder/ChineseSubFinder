package file_downloader

import (
	"crypto/sha256"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/download_file_cache"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
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

func (f *FileDownloader) Get(fromWhere string, topN int64, fileName string,
	inLanguage language.MyLanguage, fileDownloadUrl string,
	score int64, offset int64) (*supplier.SubInfo, error) {

	fileUID := fmt.Sprintf("%x", sha256.Sum256([]byte(fileDownloadUrl)))

	found, subInfo, err := f.downloadFileCache.Get(fileUID)
	if err != nil {
		return nil, err
	}
	// 如果不存在那么就先下载，然后再存入缓存中
	if found == false {
		fileData, filename, err := my_util.DownFile(f.Log, fileDownloadUrl, f.Settings.AdvancedSettings.ProxySettings)
		if err != nil {
			return nil, err
		}
		ext := ""
		if filename == "" {
			ext = filepath.Ext(fileDownloadUrl)
		} else {
			ext = filepath.Ext(filename)
		}
		inSubInfo := supplier.NewSubInfo(fromWhere, topN, fileName, inLanguage, fileDownloadUrl, score, offset, ext, fileData)
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
