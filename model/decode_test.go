package model

import (
	"testing"
)

func Test_get_IMDB_movie_xml(t *testing.T) {
    wantid := "tt0993840"
    wantyear:= "2021"
	dirPth := "x:\\电影\\Army of the Dead (2021)\\movie.xml"
	imdbInfo, err := getImdbAndYearMovieXml(dirPth)
	if err != nil {
		t.Error(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", imdbInfo.ImdbId, wantid)
	}
	if imdbInfo.Year != wantyear {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", imdbInfo.Year, wantyear)
	}
}

func Test_get_IMDB_nfo(t *testing.T) {
	wantid := "tt0993840"
	wantyear:= "2021"
	dirPth := "X:\\电影\\Army of the Dead (2021)\\Army of the Dead (2021) WEBDL-1080p.nfo"
	imdbInfo, err := getImdbAndYearNfo(dirPth)
	if err != nil {
		t.Error(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Errorf("Test_get_IMDB_movie_xml() id = %v, wantid %v", imdbInfo.ImdbId, wantid)
	}
	if imdbInfo.Year != wantyear {
		t.Errorf("Test_get_IMDB_movie_xml() year = %v, wantyear %v", imdbInfo.Year, wantyear)
	}
}

func Test_GetVideoInfoFromFileFullPath(t *testing.T) {

	subTitle := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//subTitle := "人之怒 WEBDL-1080p.mkv"
	//subTitle := "機動戦士Zガンダム WEBDL-1080p.mkv"
	//subTitle := "机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//subTitle := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	//subTitle := "X:\\连续剧\\Money.Heist\\Season 1\\Money.Heist.S01E01.SPANISH.WEBRip.x264-ION10.zh-cn.ssa"
	//subTitle := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.chi.srt"
	//subTitle := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.eng.srt"
	//subTitle := "东城梅尔 第一季第一集【YYeTs字幕组 简繁英双语字幕】Mare.of.Easttown.S01E01.Miss.Lady.Hawk.Herself.720p/1080p.AMZN.WEB-DL.DDP5.1.H.264-TEPES"
	info, modifyTime, err := GetVideoInfoFromFileFullPath(subTitle)
	if err != nil {
		t.Error(err)
	}
	println("Title:", info.Title, "Season:", info.Season, "Episode:", info.Episode, modifyTime.String())
}

func Test_GetSeasonAndEpisodeFromFileName(t *testing.T) {
	//str := `杀死伊芙 第二季(第1集-简繁英双语字幕-FIX字幕侠)Killing.Eve.S02E01.Do.You.Know.How.to.Dispose.of.a.Body.1080p.AMZN.WEB-DL.DDP5.1.H.264-NTb.rar`
	str := `杀死伊芙 第二季(-简繁英双语字幕-FIX字幕侠)Killing.Eve.S02.Do.You.Know.How.to.Dispose.of.a.Body.1080p.AMZN.WEB-DL.DDP5.1.H.264-NTb.rar`
	b, s, e, err := GetSeasonAndEpisodeFromSubFileName(str)
	if err != nil {
		t.Fatal(err)
	}
	println(b, s, e)
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