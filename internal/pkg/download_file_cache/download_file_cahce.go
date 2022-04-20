package download_file_cache

import (
	"encoding/json"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/badger_err_check"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/dgraph-io/badger/v3"
	"time"
)

type DownloadFileCache struct {
	settings *settings.Settings
}

func NewDownloadFileCache(settings *settings.Settings) *DownloadFileCache {
	return &DownloadFileCache{settings: settings}
}

func (d *DownloadFileCache) Get(fileUrlUID string) (bool, *supplier.SubInfo, error) {

	var subInfo supplier.SubInfo
	err := GetDb().View(
		func(tx *badger.Txn) error {
			var err error

			key := []byte(fileUrlUID)
			e, err := tx.Get(key)
			if err != nil {

				if badger_err_check.IsErrOk(err) == true {
					return nil
				}

				return err
			}
			valCopy, err := e.ValueCopy(nil)
			if err != nil {
				return err
			}

			err = json.Unmarshal(valCopy, &subInfo)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return false, nil, err
	}

	if subInfo.GetUID() == "" {
		return false, nil, nil
	}
	return true, &subInfo, nil
}

// Add 新增一个下载好的文件到缓存种，TTL 时间需要从 Settings 中去配置
func (d DownloadFileCache) Add(subInfo *supplier.SubInfo) error {

	err := GetDb().Update(
		func(tx *badger.Txn) error {

			var err error
			key := []byte(subInfo.GetUID())
			if err != nil {
				return err
			}

			b, err := json.Marshal(subInfo)
			if err != nil {
				return err
			}
			// 只支持秒或者小时为单位
			tmpTTL := time.Duration(d.settings.AdvancedSettings.DownloadFileCache.TTL) * time.Second
			if d.settings.AdvancedSettings.DownloadFileCache.Unit == "hour" {
				tmpTTL = time.Duration(d.settings.AdvancedSettings.DownloadFileCache.TTL) * time.Hour
			} else {
				tmpTTL = time.Duration(d.settings.AdvancedSettings.DownloadFileCache.TTL) * time.Second
			}

			e := badger.NewEntry(key, b).WithTTL(tmpTTL)
			err = tx.SetEntry(e)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}
