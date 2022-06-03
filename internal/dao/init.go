package dao

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// GetDb 获取数据库实例
func GetDb() *gorm.DB {
	if db == nil {
		once.Do(func() {
			err := InitDb()
			if err != nil {
				panic(err)
			}
		})
	}
	return db
}

// DeleteDbFile 删除 Db 文件
func DeleteDbFile() error {

	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		err = sqlDB.Close()
		if err != nil {
			return err
		}
	}

	// 这里需要考虑是 Windows 的时候就是在本程序的允许目录下新建数据库即可
	// 如果是 Linux 则在 /config 目录下
	nowDbFileName := getDbName()

	if my_util.IsFile(nowDbFileName) == true {
		return os.Remove(nowDbFileName)
	}
	return nil
}

// InitDb 初始化数据库
func InitDb() error {
	var err error
	// 新建数据库
	nowDbFileName := getDbName()

	dbDir := filepath.Dir(nowDbFileName)
	if my_util.IsDir(dbDir) == false {
		_ = os.MkdirAll(dbDir, os.ModePerm)
	}
	db, err = gorm.Open(sqlite.Open(nowDbFileName), &gorm.Config{})
	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect database, %s", err.Error()))
	}
	// 迁移 schema
	err = db.AutoMigrate(&models.HotFix{}, &models.SubFormatRec{},
		&models.IMDBInfo{}, &models.VideoSubInfo{},
		&models.ThirdPartSetVideoPlayedInfo{},
		&models.MediaInfo{},
		&models.LowVideoSubInfo{},
	)
	if err != nil {
		return errors.New(fmt.Sprintf("db AutoMigrate error, %s", err.Error()))
	}
	return nil

}

func getDbName() string {
	return filepath.Join(my_folder.GetConfigRootDirFPath(), dbFileName)
}

var (
	db   *gorm.DB
	once sync.Once
)

const (
	dbFileName = "ChineseSubFinder-Cache.db"
)
