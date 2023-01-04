package search

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
)

func TestSearchSeriesAllEpsAndSubtitles(t *testing.T) {

	seasonInfo, err := SeriesAllEpsAndSubtitles(log_helper.GetLogger4Tester(), "X:\\连续剧\\Pantheon")
	if err != nil {
		t.Fatal(err)
	}
	println(seasonInfo.Name)
}
