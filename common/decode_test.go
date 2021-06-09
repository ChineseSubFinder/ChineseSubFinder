package common

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

	movieFile1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	movieFile2 := "人之怒 WEBDL-1080p.mkv"
	movieFile3 := "機動戦士Zガンダム WEBDL-1080p.mkv"
	movieFile4 := "机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	standard1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	sub1 := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.chi.srt"
	sub2 := "Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.eng.srt"

	m1, err := GetVideoInfo(movieFile1)
	if err != nil {
		t.Error(err)
	}
	println(m1.Title, m1.Year, m1.Quality, m1.Codec, m1.Hardcoded)

	m2, err := GetVideoInfo(movieFile2)
	if err != nil {
		t.Error(err)
	}
	println(m2.Title, m2.Quality, m2.Codec, m2.Hardcoded)

	m3, err := GetVideoInfo(movieFile3)
	if err != nil {
		t.Error(err)
	}
	println(m3.Title, m3.Quality, m3.Codec, m3.Hardcoded)

	m4, err := GetVideoInfo(movieFile4)
	if err != nil {
		t.Error(err)
	}
	println(m4.Title, m4.Quality, m4.Codec, m4.Hardcoded)

	s1, err := GetVideoInfo(standard1)
	if err != nil {
		t.Error(err)
	}
	println(s1.Title, s1.Season, s1.Episode, s1.Quality, s1.Codec, s1.Hardcoded)

	osub1, err := GetVideoInfo(sub1)
	if err != nil {
		t.Error(err)
	}
	println(osub1.Title, osub1.Language)

	osub2, err := GetVideoInfo(sub2)
	if err != nil {
		t.Error(err)
	}
	println(osub2.Title, osub2.Language)
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