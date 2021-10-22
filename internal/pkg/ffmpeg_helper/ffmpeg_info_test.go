package ffmpeg_helper

import "testing"

func TestFFMPEGInfo_GetCacheFolderFPath(t *testing.T) {
	type fields struct {
		VideoFullPath    string
		AudioInfoList    []AudioInfo
		SubtitleInfoList []SubtitleInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "00", fields: fields{VideoFullPath: "c:\\abc\\xyz-asd\\xyz-asd.mp4"}, want: "c:\\abc\\xyz-asd\\" + cacheFolder + "\\xyz-asd"},
		{name: "01", fields: fields{VideoFullPath: "c:\\xyz-asd\\xyz-asd.mp4"}, want: "c:\\xyz-asd\\" + cacheFolder + "\\xyz-asd"},
		{name: "02", fields: fields{VideoFullPath: "c:\\xyz-asd.mp4"}, want: "c:\\" + cacheFolder + "\\xyz-asd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FFMPEGInfo{
				VideoFullPath:    tt.fields.VideoFullPath,
				AudioInfoList:    tt.fields.AudioInfoList,
				SubtitleInfoList: tt.fields.SubtitleInfoList,
			}
			if got := f.GetCacheFolderFPath(); got != tt.want {
				t.Errorf("GetCacheFolderFPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
