package cache_center

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"sync"
)

type CacheCenter struct {
	Settings                 *settings.Settings
	Log                      *logrus.Logger
	centerFolder             string
	downloadFileSaveRootPath string
	taskQueueSaveRootPath    string
	dbFPath                  string
	cacheName                string
	db                       *gorm.DB
	locker                   sync.Mutex
}

func NewCacheCenter(cacheName string, Settings *settings.Settings, Log *logrus.Logger) *CacheCenter {

	c := CacheCenter{}
	c.Settings = Settings
	c.Log = Log
	var err error
	c.centerFolder, err = my_folder.GetRootCacheCenterFolder()
	if err != nil {
		panic(err)
	}

	c.taskQueueSaveRootPath = filepath.Join(c.centerFolder, taskQueueFolderName, cacheName)

	c.downloadFileSaveRootPath = filepath.Join(c.centerFolder, downloadFilesFolderName, cacheName)

	c.dbFPath = filepath.Join(c.centerFolder, cacheName+"_"+dbFileName)

	c.db, err = gorm.Open(sqlite.Open(c.dbFPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database, %s", err.Error()))
	}
	// 迁移 schema
	err = c.db.AutoMigrate(&models.DailyDownloadInfo{}, &models.TaskQueueInfo{}, &models.DownloadFileInfo{})
	if err != nil {
		panic(fmt.Sprintf("db AutoMigrate error, %s", err.Error()))
	}
	return &c
}

func (c *CacheCenter) GetName() string {
	return c.cacheName
}

func (c *CacheCenter) Close() {
	defer c.locker.Unlock()
	c.locker.Lock()

	sqlDB, err := c.db.DB()
	if err != nil {
		return
	}
	err = sqlDB.Close()
	if err != nil {
		return
	}
}

func DelDb(cacheName string) {

	centerFolder, err := my_folder.GetRootCacheCenterFolder()
	if err != nil {
		return
	}
	dbFPath := filepath.Join(centerFolder, cacheName+"_"+dbFileName)
	if my_util.IsFile(dbFPath) == true {
		err = os.Remove(dbFPath)
		if err != nil {
			return
		}
	}

	taskQueueSaveRootPath := filepath.Join(centerFolder, taskQueueFolderName, cacheName)
	err = my_folder.ClearFolder(taskQueueSaveRootPath)
	if err != nil {
		return
	}

	downloadFileSaveRootPath := filepath.Join(centerFolder, downloadFilesFolderName, cacheName)
	err = my_folder.ClearFolder(downloadFileSaveRootPath)
	if err != nil {
		return
	}
}

const (
	taskQueueFolderName     = "task_queue"
	downloadFilesFolderName = "download_files"
	dbFileName              = "cache_center.db"
)
