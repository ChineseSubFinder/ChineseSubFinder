package TestCode

import "testing"

func Test_statistics_subs_score(t *testing.T) {
	type args struct {
		baseAudioFileFPath string
		baseSubFileFPath   string
		subSearchRootPath  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_statistics_subs_score",
			args: args{
				baseAudioFileFPath: "C:\\temp\\video\\base\\RM-S05E01\\未知语言_1.pcm",
				baseSubFileFPath:   "C:\\temp\\video\\base\\RM-S05E01\\英_2.srt",
				subSearchRootPath:  "X:\\电影",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statistics_subs_score(tt.args.baseAudioFileFPath, tt.args.baseSubFileFPath, tt.args.subSearchRootPath)
		})
	}
}
