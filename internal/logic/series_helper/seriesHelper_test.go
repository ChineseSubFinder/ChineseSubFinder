package series_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"testing"
)

func TestReadSeriesInfoFromDir(t *testing.T) {

	series := unit_test_helper.GetTestDataResourceRootPath([]string{"series", "Loki"}, 4, false)
	seriesInfo, err := ReadSeriesInfoFromDir(log_helper.GetLogger4Tester(), series, 90, false, false)
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
