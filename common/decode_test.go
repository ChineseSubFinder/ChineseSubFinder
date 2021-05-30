package common

import "testing"

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
		{name: "have", args: args{dirPth: "Y:\\电影\\Army of the Dead (2021)"}, want: "tt0993840", wantErr: false},
		{name: "want error", args: args{dirPth: "Y:\\电影\\"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get_IMDB_Id(tt.args.dirPth)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get_IMDB_Id() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get_IMDB_Id() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_get_IMDB_movie_xml(t *testing.T) {
    want := "tt0993840"
	dirPth := "Y:\\电影\\Army of the Dead (2021)\\movie.xml"
	got, err := get_IMDB_movie_xml(dirPth)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", got, want)
	}
}

func Test_get_IMDB_nfo(t *testing.T) {
	want := "tt0993840"
	dirPth := "Y:\\电影\\Army of the Dead (2021)\\Army of the Dead (2021) WEBDL-1080p.nfo"
	got, err := get_IMDB_nfo(dirPth)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Test_get_IMDB_movie_xml() got = %v, want %v", got, want)
	}
}