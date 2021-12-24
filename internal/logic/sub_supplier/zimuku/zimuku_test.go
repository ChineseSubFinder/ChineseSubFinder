package zimuku

import (
	series_helper2 "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"testing"
)

func TestSupplier_GetSubListFromKeyword(t *testing.T) {

	//imdbId1 := "tt3228774"
	videoName := "黑白魔女库伊拉"
	s := NewSupplier()
	outList, err := s.getSubListFromKeyword(videoName)
	if err != nil {
		t.Error(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func TestSupplier_GetSubListFromFile(t *testing.T) {
	movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\龙猫 (1988)\\龙猫 (1988) 1080p DTS.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"

	s := NewSupplier()
	outList, err := s.getSubListFromMovie(movie1)
	if err != nil {
		t.Error(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	//ser := "X:\\连续剧\\The Bad Batch" // tt12708542
	//ser := "X:\\连续剧\\杀死伊芙 (2018)"	// tt12708542
	//ser := "X:\\连续剧\\Money.Heist"
	ser := "X:\\连续剧\\黄石 (2018)"

	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper2.ReadSeriesInfoFromDir(ser, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	// 可以指定几集去调试
	epsMap := make(map[int]int, 0)
	epsMap[4] = 5
	epsMap[1] = 4
	series_helper2.SetTheSpecifiedEps2Download(seriesInfo, epsMap)

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

func TestSupplier_getSubListFromKeyword(t *testing.T) {

	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936" // 杀死伊芙
	//imdbID := "tt2990738" // 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	//imdbID := "tt6468322" 	// 纸钞屋
	//imdbID := "tt15299712" // 云南虫谷
	//imdbID := "tt3626476"  // Vacation Friends (2021)
	imdbID := "tt11192306" // Superman.and.Lois
	subhd := NewSupplier()
	subInfos, err := subhd.getSubListFromKeyword(imdbID)
	if err != nil {
		t.Fatal(err)
	}
	for i, sublist := range subInfos {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}
