package model

import (
	"testing"
)

func TestGet_IMDB_Id(t *testing.T) {
	type args struct {
		dirPth string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "have", args: args{dirPth: "x:\\电影\\Army of the Dead (2021)"}, want: "tt0993840", wantErr: false},
		{name: "want error", args: args{dirPth: "x:\\电影\\"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImdbId(tt.args.dirPth)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImdbId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetImdbId() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetIMDB_ID(t *testing.T)  {

	serPath := "X:\\连续剧\\The Bad Batch"
	id, err := GetImdbId(serPath)
	if err != nil {
		t.Fatal(err)
	}
	println(id)
}

func Test_get_IMDB_movie_xml(t *testing.T) {
    want := "tt0993840"
	dirPth := "x:\\电影\\Army of the Dead (2021)\\movie.xml"
	got, err := getImdbMovieXml(dirPth)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", got, want)
	}
}

func Test_get_IMDB_nfo(t *testing.T) {
	want := "tt0993840"
	dirPth := "X:\\电影\\Army of the Dead (2021)\\Army of the Dead (2021) WEBDL-1080p.nfo"
	got, err := getImdbNfo(dirPth)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", got, want)
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