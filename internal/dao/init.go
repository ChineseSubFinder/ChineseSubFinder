package dao

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

// InitDb 初始化数据库
func InitDb() error {
	var err error
	// 新建数据库
	db, err = gorm.Open(sqlite.Open(dbFilename), &gorm.Config{})
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
	if pkg.IsFile(dbFilename) == true {
		return os.Remove(dbFilename)
	}
	return nil
}

var (
	db *gorm.DB
)

const dbFilename = "settings.db"
