package cache_center

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center/models"
)

func (c *CacheCenter) DownloadFileAdd(subInfo *supplier.SubInfo) error {
	defer c.locker.Unlock()
	c.locker.Lock()

	if subInfo.FileUrl == "" {
		return errors.New("subInfo FileUrl is empty")
	}

	// 只支持秒或者小时为单位
	tmpTTL := time.Duration(settings.Get().AdvancedSettings.DownloadFileCache.TTL) * time.Second
	if settings.Get().AdvancedSettings.DownloadFileCache.Unit == "hour" {
		tmpTTL = time.Duration(settings.Get().AdvancedSettings.DownloadFileCache.TTL) * time.Hour
	} else {
		tmpTTL = time.Duration(settings.Get().AdvancedSettings.DownloadFileCache.TTL) * time.Second
	}

	b, err := json.Marshal(subInfo)
	if err != nil {
		return err
	}

	// 保存到本地文件
	todayString := time.Now().Format("2006-01-02")
	saveFPath := filepath.Join(c.downloadFileSaveRootPath, todayString, subInfo.GetUID())
	err = pkg.WriteFile(saveFPath, b)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(c.downloadFileSaveRootPath, saveFPath)
	if err != nil {
		return err
	}

	df := models.DownloadFileInfo{
		UID:            subInfo.GetUID(),
		RelPath:        relPath,
		ExpirationTime: time.Now().Add(tmpTTL),
	}

	c.db.Save(&df)

	return nil
}

func (c *CacheCenter) DownloadFileGet(fileUrlUID string) (bool, *supplier.SubInfo, error) {
	defer c.locker.Unlock()
	c.locker.Lock()

	var dfs []models.DownloadFileInfo
	c.db.Where("uid = ?", fileUrlUID).Find(&dfs)

	if len(dfs) == 0 {
		return false, nil, nil
	}

	localFileFPath := filepath.Join(c.downloadFileSaveRootPath, dfs[0].RelPath)
	if pkg.IsFile(localFileFPath) == false {
		return false, nil, nil
	}

	bytes, err := os.ReadFile(localFileFPath)
	if err != nil {
		return false, nil, err
	}

	var subInfo supplier.SubInfo
	err = json.Unmarshal(bytes, &subInfo)
	if err != nil {
		return false, nil, err
	}

	return true, &subInfo, nil
}
