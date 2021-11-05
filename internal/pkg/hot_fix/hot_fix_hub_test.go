package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"testing"
)

func TestHotFixProcess(t *testing.T) {

	// 先删除 db
	err := dao.DeleteDbFile()
	if err != nil {
		t.Fatal(err)
	}
	testDataPath := "../../../TestData/hotfix/001"
	movieDir := "movies"
	seriesDir := "series"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	// 测试文件夹
	testMovieDir := filepath.Join(testRootDir, movieDir)
	testSeriesDir := filepath.Join(testRootDir, seriesDir)
	// 开始修复
	err = HotFixProcess(types.HotFixParam{
		MovieRootDir:  testMovieDir,
		SeriesRootDir: testSeriesDir,
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
