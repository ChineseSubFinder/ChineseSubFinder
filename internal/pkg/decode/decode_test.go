package decode

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetImdbAndYearMovieXml(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "Army of the Dead (2021)"}, 4, false)

	wantid := "tt0993840"
	wantyear := "2021"
	dirPth := filepath.Join(rootDir, "movie.xml")
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

func Test_getImdbAndYearNfo(t *testing.T) {
	type args struct {
		nfoFilePath string
		rootKey     string
	}
	tests := []struct {
		name    string
		args    args
		want    types.VideoIMDBInfo
		wantErr bool
	}{
		{
			name: "Army of the Dead (2021) WEBDL-1080p.nfo", args: args{
				nfoFilePath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "Army of the Dead (2021)"}, 4, false), "Army of the Dead (2021) WEBDL-1080p.nfo"),
				rootKey:     "movie",
			},
			want: types.VideoIMDBInfo{
				ImdbId:      "tt0993840",
				Title:       "活死人军团",
				Year:        "2021",
				ReleaseDate: "2021-05-13",
			},
			wantErr: false,
		},
		{
			name: "tvshow_00 (2).nfo", args: args{
				nfoFilePath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"nfo_files", "tvshow"}, 4, false), "tvshow_00 (2).nfo"),
				rootKey:     "tvshow",
			},
			want: types.VideoIMDBInfo{
				ImdbId:      "tt0346314",
				Title:       "Ghost in the Shell: Stand Alone Complex",
				ReleaseDate: "2002-10-01",
			},
			wantErr: false,
		},
		{
			name: "tvshow_00 (3).nfo", args: args{
				nfoFilePath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"nfo_files", "tvshow"}, 4, false), "tvshow_00 (3).nfo"),
				rootKey:     "tvshow",
			},
			want: types.VideoIMDBInfo{
				ImdbId:      "tt1856010",
				Title:       "House of Cards (US)",
				ReleaseDate: "2013-02-01",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getImdbAndYearNfo(tt.args.nfoFilePath, tt.args.rootKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("getImdbAndYearNfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImdbAndYearNfo() got = %v, want %v", got, tt.want)
			}
		})
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
				ImdbId:      "tt0096697",
				Title:       "The Simpsons",
				ReleaseDate: "1989-12-17",
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

func TestIsFakeBDMVWorked(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "失控玩家 (2021)"}, 4, false)

	bok, dbmvFPath := IsFakeBDMVWorked(filepath.Join(rootDir, "失控玩家 (2021).mp4"))
	if bok == false {
		t.Fatal("IsFakeBDMVWorked error")
	}
	println(dbmvFPath)
}
