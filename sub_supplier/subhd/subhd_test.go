package subhd

import (
	"github.com/allanpk716/ChineseSubFinder/series_helper"
	"testing"
)

func TestSupplier_GetSubListFromFile(t *testing.T) {

	//movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	movie1 := "X:\\电影\\Luca (2021)\\Luca (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\Oslo (2021)\\Oslo (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	shooter := NewSupplier()
	outList, err := shooter.getSubListFromFile4Movie(movie1)
	if err != nil {
		t.Error(err)
	}
	println(outList)

	if len(outList) == 0 {
		println("now sub found")
	}

	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, sublist.FileUrl, len(sublist.Data))
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	ser := "X:\\连续剧\\The Bad Batch"	// tt12708542
	//ser := "X:\\连续剧\\杀死伊芙 (2018)"	// tt7016936
	//ser := "X:\\连续剧\\Money.Heist"
	//ser := "X:\\连续剧\\黑钱胜地 (2017)"


	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(ser, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := NewSupplier()
	outList, err := s.GetSubListFromFile4Series(seriesInfo)
	if err != nil {
		t.Fatal(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}