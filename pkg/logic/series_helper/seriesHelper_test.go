package series_helper

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestReadSeriesInfoFromDir(t *testing.T) {

	//series := unit_test_helper.GetTestDataResourceRootPath([]string{"series", "Loki"}, 4, false)
	seriesInfo, err := ReadSeriesInfoFromDir(log_helper.GetLogger4Tester(), "X:\\连续剧\\黑袍纠察队 (2019)", 90, false, true)
	if err != nil {
		t.Fatal(err)
	}

	println(seriesInfo.Name, seriesInfo.Year, seriesInfo.ImdbId)
	for i, info := range seriesInfo.EpList {
		println("Video:", i, info.Season, info.Episode)
		for j, subInfo := range info.SubAlreadyDownloadedList {
			println("Sub:", j, subInfo.Title, subInfo.Season, subInfo.Episode, subInfo.Language.String())
		}
	}
}

func TestGetSeriesListFromDirs(t *testing.T) {

	series := unit_test_helper.GetTestDataResourceRootPath([]string{"series"}, 4, false)
	got, err := GetSeriesListFromDirs(log_helper.GetLogger4Tester(), []string{series})
	if err != nil {
		t.Fatal(err)
	}

	if got.Size() < 1 {
		t.Fatal("GetSeriesListFromDirs got len < 1")
	}
}
