package dao

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sqlite"
	"gorm.io/gorm"
	"os"
	"runtime"
)

// InitDb 初始化数据库
func InitDb() error {
	var err error
	// 新建数据库
	nowDbFileName := getDbName()
	db, err = gorm.Open(sqlite.Open(nowDbFileName), &gorm.Config{})
	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect database, %s", err.Error()))
	}
	// 迁移 schema
	err = db.AutoMigrate(&models.HotFix{})
	if err != nil {
		return errors.New(fmt.Sprintf("db AutoMigrate error, %s", err.Error()))
	}
	return nil

}

// GetDb 获取数据库实例
func GetDb() *gorm.DB {
	return db
}

// DeleteDbFile 删除 Db 文件
func DeleteDbFile() error {

	// 这里需要考虑是 Windows 的时候就是在本程序的允许目录下新建数据库即可
	// 如果是 Linux 则在 /config 目录下
	nowDbFileName := getDbName()

	if pkg.IsFile(nowDbFileName) == true {
		return os.Remove(nowDbFileName)
	}
	return nil
}

func getDbName() string {
	nowDbFileName := ""
	sysType := runtime.GOOS
	if sysType == "linux" {
		nowDbFileName = dbFileNameLinux
	}
	if sysType == "windows" {
		nowDbFileName = dbFileNameWindows
	}
	return nowDbFileName
}

var (
	db *gorm.DB
)

const (
	dbFileNameLinux   = "/config/settings.db"
	dbFileNameWindows = "settings.db"
)
