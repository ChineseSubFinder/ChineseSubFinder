package model

import (
	"testing"
)

func Test_GetIMDB_ID(t *testing.T)  {

	serPath := "X:\\连续剧\\The Bad Batch"
	id, year, err := GetImdbIdAndYear(serPath)
	if err != nil {
		t.Fatal(err)
	}
	println(id, year)
}

func Test_get_IMDB_movie_xml(t *testing.T) {
    wantid := "tt0993840"
    wantyear:= "2021"
	dirPth := "x:\\电影\\Army of the Dead (2021)\\movie.xml"
	id, year, err := getImdbAndYearMovieXml(dirPth)
	if err != nil {
		t.Error(err)
	}
	if id != wantid {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", id, wantid)
	}
	if year != wantyear {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", year, wantyear)
	}
}

func Test_get_IMDB_nfo(t *testing.T) {
	wantid := "tt0993840"
	wantyear:= "2021"
	dirPth := "X:\\电影\\Army of the Dead (2021)\\Army of the Dead (2021) WEBDL-1080p.nfo"
	id, year, err := getImdbAndYearNfo(dirPth)
	if err != nil {
		t.Error(err)
	}
	if id != wantid {
		t.Errorf("Test_get_IMDB_movie_xml() id = %v, wantid %v", id, wantid)
	}
	if year != wantyear {
		t.Errorf("Test_get_IMDB_movie_xml() year = %v, wantyear %v", id, wantyear)
	}
}

func Test_VideoInfo(t *testing.T) {

	//subTitle := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//subTitle := "人之怒 WEBDL-1080p.mkv"
	//subTitle := "機動戦士Zガンダム WEBDL-1080p.mkv"
	//subTitle := "机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//subTitle := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	//subTitle := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.chi.srt"
	//subTitle := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.eng.srt"
	subTitle := "东城梅尔 第一季第一集【YYeTs字幕组 简繁英双语字幕】Mare.of.Easttown.S01E01.Miss.Lady.Hawk.Herself.720p/1080p.AMZN.WEB-DL.DDP5.1.H.264-TEPES"
	info, err := GetVideoInfo(subTitle)
	if err != nil {
		t.Error(err)
	}
	println("Title:", info.Title, "Season:", info.Season, "Episode:", info.Episode)
}

func TestGetNumber2Float(t *testing.T) {
	testString := "asd&^%1998.2jh aweo "
	outNumber, err := GetNumber2Float(testString)
	if err != nil {
		t.Error(err)
	}
	if outNumber != 1998.2 {
		t.Error("not the same")
	}
}

func TestGetNumber2int(t *testing.T) {

	testString := "asd&^%1998jh aweo "
	outNumber, err := GetNumber2int(testString)
	if err != nil {
		t.Error(err)
	}
	if outNumber != 1998 {
		t.Error("not the same")
	}
}