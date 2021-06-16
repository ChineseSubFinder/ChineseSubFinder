package series_helper

import (
	"testing"
)

func TestReadSeriesInfoFromDir(t *testing.T) {

	series := "X:\\连续剧\\杀死伊芙 (2018)"
	//series := "X:\\连续剧\\Money.Heist"

	seriesInfo, err := ReadSeriesInfoFromDir(series)
	if err != nil {
		t.Fatal(err)
	}

	println(seriesInfo.Name, seriesInfo.Year, seriesInfo.ImdbId)
	for i, info := range seriesInfo.EpList {
		println("Video:", i, info.Season, info.Episode)
		for j, subInfo := range info.SubList {
			println("Sub:", j, subInfo.Title, subInfo.Season, subInfo.Episode, subInfo.Language.String())
		}
	}
}
