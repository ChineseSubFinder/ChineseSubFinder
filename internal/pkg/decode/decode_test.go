package decode

import (
	"fmt"
	"testing"
    "path/filepath"
)

func TestGetImdbAndYearMovieXml(t *testing.T) {
	wantid := "tt0993840"
	wantyear := "2021"
	dirPth := filepath.FromSlash("../../../TestData/video_info_file/movie.xml")
	imdbInfo, err := getImdbAndYearMovieXml(dirPth)
	if err != nil {
		t.Error(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Errorf("got = %v, want %v", imdbInfo.ImdbId, wantid)
	}
	if imdbInfo.Year != wantyear {
		t.Errorf("got = %v, want %v", imdbInfo.Year, wantyear)
	}
}

func TestGetImdbAndYearNfo(t *testing.T) {
	wantid := "tt0993840"
	wantyear := "2021"
	dirPth := filepath.FromSlash("../../../TestData/video_info_file/Army of the Dead (2021) WEBDL-1080p.nfo")
	imdbInfo, err := getImdbAndYearNfo(dirPth, "movie")
	if err != nil {
		t.Fatal(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Fatal(fmt.Sprintf("id = %v, wantid %v", imdbInfo.ImdbId, wantid))
	}
	if imdbInfo.Year != wantyear {
		t.Fatal(fmt.Sprintf("year = %v, wantyear %v", imdbInfo.Year, wantyear))
	}

	wantid = "tt12801326"
	wantyear = "2020"
	dirPth = filepath.FromSlash("../../../TestData/video_info_file/has_http_address.nfo")
	imdbInfo, err = getImdbAndYearNfo(dirPth, "movie")
	if err != nil {
		t.Fatal(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Fatal(fmt.Sprintf("id = %v, wantid %v", imdbInfo.ImdbId, wantid))
	}
	if imdbInfo.Year != wantyear {
		t.Fatal(fmt.Sprintf("year = %v, wantyear %v", imdbInfo.Year, wantyear))
	}

	wantid = ""
	wantyear = ""
	dirPth = filepath.FromSlash("../../../TestData/video_info_file/only_http_address.nfo")
	imdbInfo, err = getImdbAndYearNfo(dirPth, "movie")
	if err == nil {
		t.Fatal("need errot")
	}
}

func TestGetVideoInfoFromFileFullPath(t *testing.T) {

	subTitle := filepath.FromSlash("../../../TestData/video_info_file/Friends (1994) - s08e02.mp4")

	info, modifyTime, err := GetVideoInfoFromFileFullPath(subTitle)
	if err != nil || info.Season != 8 || info.Episode != 2 {
		t.Error(err)
	}
	t.Logf("\n\nTitle: %s Season: %d Episode: %d Modified Time: %s" , info.Title, info.Season, info.Episode, modifyTime.String())
}

func TestGetSeasonAndEpisodeFromFileName(t *testing.T) {
	str := `杀死伊芙 第二季(-简繁英双语字幕-FIX字幕侠)Killing.Eve.S02.Do.You.Know.How.to.Dispose.of.a.Body.1080p.AMZN.WEB-DL.DDP5.1.H.264-NTb.rar`
	b, s, e, err := GetSeasonAndEpisodeFromSubFileName(str)
	if err != nil {
		t.Fatal(err)
	}
    t.Logf("\n\n%t\t S%dE%d\n", b, s, e)
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

func Test_getImdbAndYearNfo(t *testing.T) {

	nfoInfo := filepath.FromSlash("../../../TestData/video_info_file/Army of the Dead (2021) WEBDL-1080p.nfo")
	nfo, err := getImdbAndYearNfo(nfoInfo, "movie")
    t.Logf("\n\nMovies:\timdbid\tYear\tReleaseDate\n" + 
			"        %s\t%s\t %s\n", nfo.ImdbId, nfo.Year, nfo.ReleaseDate)
	if err != nil {
		t.Fatal(err)
	}
}
