package decode

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"reflect"
	"testing"
)

func getTestFileDir(testFileName string) (xmlDir string) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "Army of the Dead (2021)"}, 4, false)

	if testFileName == "movie.xml" {
		return filepath.Join(rootDir, "movie.xml")
	} else if testFileName == "movie.nfo" {
		return filepath.Join(rootDir, "Army of the Dead (2021) WEBDL-1080p.nfo")
	} else if testFileName == "has_http_address.nfo" {
		return filepath.Join(rootDir, "has_http_address.nfo")
	} else if testFileName == "has_http_address.nfo" {
		return filepath.Join(rootDir, "only_http_address.nfo")
	}

	return ""
}

func TestGetImdbAndYearMovieXml(t *testing.T) {
	wantid := "tt0993840"
	wantyear := "2021"
	dirPth := getTestFileDir("movie.xml")
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

	imdbInfo, err := getImdbAndYearNfo(getTestFileDir("movie.nfo"), "movie")
	if err != nil {
		t.Fatal(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Fatalf("\n\nid = %v, wantid %v", imdbInfo.ImdbId, wantid)
	}
	if imdbInfo.Year != wantyear {
		t.Fatalf("\n\nyear = %v, wantyear %v", imdbInfo.Year, wantyear)
	}

	wantid = "tt12801326"
	wantyear = "2020"
	dirPth := getTestFileDir("has_http_address.nfo")
	imdbInfo, err = getImdbAndYearNfo(dirPth, "movie")
	if err != nil {
		t.Fatal(err)
	}
	if imdbInfo.ImdbId != wantid {
		t.Fatalf("\n\nid = %v, wantid %v", imdbInfo.ImdbId, wantid)
	}
	if imdbInfo.Year != wantyear {
		t.Fatalf("\n\nyear = %v, wantyear %v", imdbInfo.Year, wantyear)
	}

	wantid = ""
	wantyear = ""
	dirPth = getTestFileDir("only_http_address.nfo")
	imdbInfo, err = getImdbAndYearNfo(dirPth, "movie")
	if err == nil {
		t.Fatal("need error")
	}
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

func TestGetImdbInfo4SeriesDir(t *testing.T) {

	type args struct {
		seriesDir string
	}
	tests := []struct {
		name    string
		args    args
		want    types.VideoIMDBInfo
		wantErr bool
	}{
		{
			name: "Loki",
			args: args{seriesDir: unit_test_helper.GetTestDataResourceRootPath([]string{"series", "Loki"}, 4, false)},
			want: types.VideoIMDBInfo{
				ImdbId:      "tt9140554",
				Title:       "Loki",
				ReleaseDate: "2021-06-09",
			},
			wantErr: false,
		},
		{
			name: "辛普森一家",
			args: args{seriesDir: unit_test_helper.GetTestDataResourceRootPath([]string{"series", "辛普森一家"}, 4, false)},
			want: types.VideoIMDBInfo{
				ImdbId:      "tt9140554",
				Title:       "Loki",
				ReleaseDate: "2021-06-09",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImdbInfo4SeriesDir(tt.args.seriesDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImdbInfo4SeriesDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImdbInfo4SeriesDir() got = %v, want %v", got, tt.want)
			}
		})
	}
}
