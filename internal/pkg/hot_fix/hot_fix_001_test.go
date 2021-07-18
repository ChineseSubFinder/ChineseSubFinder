package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"testing"
)

func TestHotFix001_GetKey(t *testing.T) {
	hf001 := NewHotFix001("", "")
	if hf001.GetKey() != "001" {
		t.Fatal("GetKey() != 001")
	}
}

func TestHotFix001_Process(t *testing.T) {
	testDataPath := "..\\..\\..\\TestData\\hotfix\\001"
	movieDir := "movies"
	seriesDir := "series"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		return
	}
	// 测试文件夹
	testMovieDir := path.Join(testRootDir, movieDir)
	testSeriesDir := path.Join(testRootDir, seriesDir)
	// 开始修复
	hf001 := NewHotFix001(testMovieDir, testSeriesDir)
	err = hf001.Process()
	if err != nil {
		t.Fatal("Process ", err.Error())
	}
}
