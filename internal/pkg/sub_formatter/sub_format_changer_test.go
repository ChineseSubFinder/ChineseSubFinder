package sub_formatter

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"path"
	"testing"
)

func TestSubFormatChanger_AutoDetectThenChangeTo(t *testing.T) {

	movie_name := "AAA_org_emby"
	series_name := "Loki"
	testDataPath := "../../../TestData/sub_format_changer"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	movieDir := path.Join(testRootDir, "movie")
	seriesDir := path.Join(testRootDir, "series")

	s := NewSubFormatChanger(movieDir, seriesDir)

	movieOneDir := path.Join(movieDir, movie_name)
	seriesOneDir := path.Join(seriesDir, series_name, "Season 1")

	type args struct {
		desFormatter common.FormatterName
	}
	tests := []struct {
		name    string
		args    args
		want    RenameResults
		wantErr bool
	}{
		{name: "movie", args: args{desFormatter: common.Normal}, want: RenameResults{
			RenamedFiles: map[string]int{
				path.Join(movieOneDir, "AAA.zh.ass"):                    2,
				path.Join(movieOneDir, "AAA.zh.default.ass"):            1,
				path.Join(movieOneDir, "AAA.zh.srt"):                    1,
				path.Join(seriesOneDir, "Loki - S01E01.zh.ass"):         5,
				path.Join(seriesOneDir, "Loki - S01E01.zh.default.ass"): 1,
				path.Join(seriesOneDir, "Loki - S01E01.zh.srt"):         1,
			},
		}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
