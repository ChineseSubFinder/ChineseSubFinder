package sub_formatter

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestSubFormatChanger_AutoDetectThenChangeTo(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_format_changer"}, 4, true)
	movie_name := "AAA"
	series_name := "Loki"

	// Emby 的信息
	movieDir_org_emby := filepath.Join(testRootDir, "movie_org_emby")
	seriesDir_org_emby := filepath.Join(testRootDir, "series_org_emby")
	movieOneDir_org_emby := filepath.Join(movieDir_org_emby, movie_name)
	seriesOneDir_org_emby := filepath.Join(seriesDir_org_emby, series_name, "Season 1")
	// Normal 的信息
	movieDir_org_normal := filepath.Join(testRootDir, "movie_org_normal")
	seriesDir_org_normal := filepath.Join(testRootDir, "series_org_normal")
	movieOneDir_org_normal := filepath.Join(movieDir_org_normal, movie_name)
	seriesOneDir_org_normal := filepath.Join(seriesDir_org_normal, series_name, "Season 1")
	// emby 转 emby 理论上不应该改文件
	movieDir_emby_2_emby := filepath.Join(testRootDir, "movie_emby_2_emby")
	seriesDir_emby_2_emby := filepath.Join(testRootDir, "series_emby_2_emby")

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
					filepath.Join(movieOneDir_org_emby, "AAA.zh.ass"):                    2,
					filepath.Join(movieOneDir_org_emby, "AAA.zh.default.ass"):            1,
					filepath.Join(movieOneDir_org_emby, "AAA.zh.srt"):                    1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.ass"):         5,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.default.ass"): 1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.srt"):         1,
				},
			}, wantErr: false},
		{name: "normal 2 emby",
			fields: fields{movieRootDir: movieDir_org_normal, seriesRootDir: seriesDir_org_normal},
			args:   args{desFormatter: common.Emby},
			want: RenameResults{
				RenamedFiles: map[string]int{
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).ass"):                    1,
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).default.ass"):            1,
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).srt"):                    1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(繁英).ass"):         1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).default.ass"): 1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).srt"):         1,
				},
			}, wantErr: false},
		{name: "emby 2 emby",
			fields:  fields{movieRootDir: movieDir_emby_2_emby, seriesRootDir: seriesDir_emby_2_emby},
			args:    args{desFormatter: common.Emby},
			want:    RenameResults{},
			wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := NewSubFormatChanger([]string{tt.fields.movieRootDir}, []string{tt.fields.seriesRootDir})

			got, err := s.AutoDetectThenChangeTo(tt.args.desFormatter)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoDetectThenChangeTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.ErrFiles) > 0 {
				t.Errorf("AutoDetectThenChangeTo() got.ErrFiles len > 0")
				return
			}

			//for s2, i := range tt.want.RenamedFiles {
			//	println(s2, i)
			//}
			//println("-------------------------------")
			//for s2, i := range got.RenamedFiles {
			//	println(s2, i)
			//}
			for fileName, counter := range got.RenamedFiles {
				if tt.want.RenamedFiles[filepath.FromSlash(fileName)] != counter {
					//println(fileName)
					//println(filepath.FromSlash(fileName))
					t.Errorf("AutoDetectThenChangeTo() RenamedFiles %v got = %v, want %v", fileName, counter, tt.want.RenamedFiles[fileName])
					return
				}
			}
		})
	}
}

func TestSubFormatChangerProcess(t *testing.T) {

	// 先删除 db
	err := dao.DeleteDbFile()
	if err != nil {
		t.Fatal(err)
	}

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_format_changer"}, 4, true)
	movie_name := "AAA"
	series_name := "Loki"

	// Emby 的信息
	movieDir_org_emby := filepath.Join(testRootDir, "movie_org_emby")
	seriesDir_org_emby := filepath.Join(testRootDir, "series_org_emby")
	movieOneDir_org_emby := filepath.Join(movieDir_org_emby, movie_name)
	seriesOneDir_org_emby := filepath.Join(seriesDir_org_emby, series_name, "Season 1")
	// Normal 的信息
	movieDir_org_normal := filepath.Join(testRootDir, "movie_org_normal")
	seriesDir_org_normal := filepath.Join(testRootDir, "series_org_normal")
	movieOneDir_org_normal := filepath.Join(movieDir_org_normal, movie_name)
	seriesOneDir_org_normal := filepath.Join(seriesDir_org_normal, series_name, "Season 1")
	// emby 转 emby 理论上不应该改文件
	movieDir_emby_2_emby := filepath.Join(testRootDir, "movie_emby_2_emby")
	seriesDir_emby_2_emby := filepath.Join(testRootDir, "series_emby_2_emby")

	type args struct {
		movieRootDir    string
		seriesRootDir   string
		nowDesFormatter common.FormatterName
	}
	tests := []struct {
		name    string
		args    args
		want    RenameResults
		wantErr bool
	}{
		// 先从 emby 2 normal
		{name: "emby 2 normal",
			args: args{movieRootDir: movieDir_org_emby, seriesRootDir: seriesDir_org_emby, nowDesFormatter: common.Normal},
			want: RenameResults{
				RenamedFiles: map[string]int{
					filepath.Join(movieOneDir_org_emby, "AAA.zh.ass"):                    2,
					filepath.Join(movieOneDir_org_emby, "AAA.zh.default.ass"):            1,
					filepath.Join(movieOneDir_org_emby, "AAA.zh.srt"):                    1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.ass"):         5,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.default.ass"): 1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.zh.srt"):         1,
				},
			}, wantErr: false},
		// 然后从上面一个测试用例的文件夹中，继续，转 normal 2 emby
		{name: "normal 2 emby",
			args: args{movieRootDir: movieDir_org_emby, seriesRootDir: movieDir_org_emby, nowDesFormatter: common.Emby},
			want: RenameResults{
				RenamedFiles: map[string]int{
					filepath.Join(movieOneDir_org_emby, "AAA.chinese(简英).ass"):                    1,
					filepath.Join(movieOneDir_org_emby, "AAA.chinese(简英).default.ass"):            1,
					filepath.Join(movieOneDir_org_emby, "AAA.chinese(简英).srt"):                    1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.chinese(繁英).ass"):         1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.chinese(简英).default.ass"): 1,
					filepath.Join(seriesOneDir_org_emby, "Loki - S01E01.chinese(简英).srt"):         1,
				},
			}, wantErr: false},

		{name: "emby 2 emby",
			args:    args{movieRootDir: movieDir_emby_2_emby, seriesRootDir: seriesDir_emby_2_emby, nowDesFormatter: common.Emby},
			want:    RenameResults{},
			wantErr: false},

		// 重新评估 normal 2 emby ，需要清理数据库
		{name: "normal 2 emby new",
			args: args{movieRootDir: movieDir_org_normal, seriesRootDir: seriesDir_org_normal, nowDesFormatter: common.Emby},
			want: RenameResults{
				RenamedFiles: map[string]int{
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).ass"):                    1,
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).default.ass"):            1,
					filepath.Join(movieOneDir_org_normal, "AAA.chinese(简英).srt"):                    1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(繁英).ass"):         1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).default.ass"): 1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.chinese(简英).srt"):         1,
				},
			}, wantErr: false},
		// 然后从上面一个测试用例的文件夹中，继续，转 emby 2 normal
		{name: "emby 2 normal new",
			args: args{movieRootDir: movieDir_org_normal, seriesRootDir: seriesDir_org_normal, nowDesFormatter: common.Normal},
			want: RenameResults{
				RenamedFiles: map[string]int{
					filepath.Join(movieOneDir_org_normal, "AAA.zh.ass"):                    1,
					filepath.Join(movieOneDir_org_normal, "AAA.zh.default.ass"):            1,
					filepath.Join(movieOneDir_org_normal, "AAA.zh.srt"):                    1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.zh.ass"):         1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.zh.default.ass"): 1,
					filepath.Join(seriesOneDir_org_normal, "Loki - S01E01.zh.srt"):         1,
				},
			}, wantErr: false},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if i == 0 || i == 2 || i == 3 {
				// 0 - 1 轮次，测试的是 先从 emby 2 normal
				// 然后从上面一个测试用例的文件夹中，继续，转 normal 2 emby
				// 先删除 db
				err = dao.DeleteDbFile()
				if err != nil {
					t.Fatal(err)
				}
				err = dao.InitDb()
				if err != nil {
					t.Fatal(err)
				}
			}

			got, err := SubFormatChangerProcess([]string{tt.args.movieRootDir}, []string{tt.args.seriesRootDir}, tt.args.nowDesFormatter)
			if err != nil != tt.wantErr {
				t.Errorf("SubFormatChangerProcess() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got.ErrFiles) > 0 {
				t.Errorf("SubFormatChangerProcess() got.ErrFiles len > 0")
				return
			}

			for s2, i := range tt.want.RenamedFiles {
				println(s2, i)
			}
			println("-------------------------------")
			for s2, i := range got.RenamedFiles {
				println(s2, i)
			}

			for fileName, counter := range got.RenamedFiles {
				if tt.want.RenamedFiles[filepath.FromSlash(fileName)] != counter {
					//println(fileName)
					//println(filepath.FromSlash(fileName))
					t.Errorf("SubFormatChangerProcess() RenamedFiles %v got = %v, want %v", fileName, counter, tt.want.RenamedFiles[fileName])
					return
				}
			}

			if i == 0 {
				// 这里需要校验一次数据库的赋值是否正确
				var subFormatRec models.SubFormatRec
				dao.GetDb().First(&subFormatRec)
				if subFormatRec.FormatName != int(common.Normal) || subFormatRec.Done == false {
					t.Fatal(tt.name, "i == 0 check db result")
				}
			}
			if i == 1 {
				// 这里需要校验一次数据库的赋值是否正确
				var subFormatRec models.SubFormatRec
				dao.GetDb().First(&subFormatRec)
				if subFormatRec.FormatName != int(common.Emby) || subFormatRec.Done == false {
					t.Fatal(tt.name, "i == 1 check db result")
				}
			}

			if i == 3 {
				// 这里需要校验一次数据库的赋值是否正确
				var subFormatRec models.SubFormatRec
				dao.GetDb().First(&subFormatRec)
				if subFormatRec.FormatName != int(common.Emby) || subFormatRec.Done == false {
					t.Fatal(tt.name, "i == 3 check db result")
				}
			}

			if i == 4 {
				// 这里需要校验一次数据库的赋值是否正确
				var subFormatRec models.SubFormatRec
				dao.GetDb().First(&subFormatRec)
				if subFormatRec.FormatName != int(common.Normal) || subFormatRec.Done == false {
					t.Fatal(tt.name, "i == 4 check db result")
				}
			}
		})
	}
}
