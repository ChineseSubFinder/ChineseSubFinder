package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"testing"
)

func TestHotFixProcess(t *testing.T) {

	// 先删除 db
	err := dao.DeleteDbFile()
	if err != nil {
		t.Fatal(err)
	}
	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"hotfix", "001"}, 4, true)
	movieDir := "movies"
	seriesDir := "series"
	// 测试文件夹
	testMovieDir := filepath.Join(testDataPath, movieDir)
	testSeriesDir := filepath.Join(testDataPath, seriesDir)
	// 开始修复
	err = HotFixProcess(log_helper.GetLogger4Tester(), types.HotFixParam{
		MovieRootDirs:  []string{testMovieDir},
		SeriesRootDirs: []string{testSeriesDir},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 判断数据库的标记是否正确
	hotfixResult := models.HotFix{}
	result := dao.GetDb().Where("key = ?", "001").First(&hotfixResult)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
