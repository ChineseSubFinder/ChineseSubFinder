package sub_formatter

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"path"
	"testing"
)

func TestSubFormatChanger_AutoDetectThenChangeTo(t *testing.T) {

	testDataPath := "../../../TestData/sub_format_changer"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	movie_name := "AAA"
	series_name := "Loki"

	// Emby 的信息
	movieDir_org_emby := path.Join(testRootDir, "movie_org_emby")
	seriesDir_org_emby := path.Join(testRootDir, "series_org_emby")
	movieOneDir_org_emby := path.Join(movieDir_org_emby, movie_name)
	seriesOneDir_org_emby := path.Join(seriesDir_org_emby, series_name, "Season 1")
	// Normal 的信息
	movieDir_org_normal := path.Join(testRootDir, "movie_org_normal")
	seriesDir_org_normal := path.Join(testRootDir, "series_org_normal")
	movieOneDir_org_normal := path.Join(movieDir_org_normal, movie_name)
	seriesOneDir_org_normal := path.Join(seriesDir_org_normal, series_name, "Season 1")

	type fields struct {
		movieRootDir  string
		seriesRootDir string
	}
	type args struct {
		desFormatter common.FormatterName
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    RenameResults
		wantErr bool
	}{
		{name: "emby 2 normal",
			fields: fields{movieRootDir: movieDir_org_emby, seriesRootDir: seriesDir_org_emby},
			args:   args{desFormatter: common.Normal},
			want: RenameResults{
				RenamedFiles: map[string]int{
					path.Join(movieOneDir_org_emby, "AAA.zh.ass"):                    2,
					path.Join(movieOneDir_org_emby, "AAA.zh.default.ass"):            1,
					path.Join(movieOneDir_org_emby, "AAA.zh.srt"):                    1,
					path.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.ass"):         5,
					path.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.default.ass"): 1,
					path.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.srt"):         1,
				},
			}, wantErr: false},
		{name: "normal 2 emby",
			fields: fields{movieRootDir: movieDir_org_normal, seriesRootDir: seriesDir_org_normal},
			args:   args{desFormatter: common.Emby},
			want: RenameResults{
				RenamedFiles: map[string]int{
					path.Join(movieOneDir_org_normal, "AAA.chinese(简英).ass"):                    1,
					path.Join(movieOneDir_org_normal, "AAA.chinese(简英).default.ass"):            1,
					path.Join(movieOneDir_org_normal, "AAA.chinese(简英).srt"):                    1,
					path.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(繁英).ass"):         1,
					path.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).default.ass"): 1,
					path.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).srt"):         1,
				},
			}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := NewSubFormatChanger(tt.fields.movieRootDir, tt.fields.seriesRootDir)

			got, err := s.AutoDetectThenChangeTo(tt.args.desFormatter)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoDetectThenChangeTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.ErrFiles) > 0 {
				t.Errorf("AutoDetectThenChangeTo() got.ErrFiles len > 0")
				return
			}

			for fileName, counter := range got.RenamedFiles {
				if tt.want.RenamedFiles[fileName] != counter {
					t.Errorf("AutoDetectThenChangeTo() RenamedFiles %v got = %v, want %v", fileName, counter, tt.want.RenamedFiles[fileName])
					return
				}
			}
		})
	}
}
