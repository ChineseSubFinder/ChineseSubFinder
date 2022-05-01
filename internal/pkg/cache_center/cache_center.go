package cache_center

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"sync"
)

type CacheCenter struct {
	Settings                 *settings.Settings
	Log                      *logrus.Logger
	centerFolder             string
	downloadFileSaveRootPath string
	db                       *gorm.DB
	locker                   sync.Mutex
}

func NewCacheCenter(Settings *settings.Settings, Log *logrus.Logger) *CacheCenter {

	c := CacheCenter{}
	c.Settings = Settings
	c.Log = Log
	var err error
	c.centerFolder, err = my_folder.GetRootCacheCenterFolder()
	if err != nil {
		panic(err)
	}
	c.downloadFileSaveRootPath = filepath.Join(c.centerFolder, "download_files")
	dbFPath := filepath.Join(c.centerFolder, dbFileName)
	c.db, err = gorm.Open(sqlite.Open(dbFPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database, %s", err.Error()))
	}
	// 迁移 schema
	err = c.db.AutoMigrate(&models.DailyDownloadInfo{}, &models.DailyDownloadInfo{})
	if err != nil {
		panic(fmt.Sprintf("db AutoMigrate error, %s", err.Error()))
	}
	return &c
}

const (
	dbFileName = "cache_center.db"
)
