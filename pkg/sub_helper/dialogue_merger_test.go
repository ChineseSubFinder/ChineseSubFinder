package sub_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"

	// "github.com/ChineseSubFinder/ChineseSubFinder/pkg/my_util"
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
)

func Test_isFirstLetterIsEngUpper(t *testing.T) {
	type args struct {
		instring string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "0", args: args{instring: "A"}, want: true},
		{name: "1", args: args{instring: "a"}, want: false},
		{name: "2", args: args{instring: "哈"}, want: false},
		{name: "3", args: args{instring: ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFirstLetterIsEngUpper(tt.args.instring); got != tt.want {
				t.Errorf("isFirstLetterIsEngUpper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isFirstLetterIsEngLower(t *testing.T) {
	type args struct {
		instring string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "0", args: args{instring: "A"}, want: false},
		{name: "1", args: args{instring: "a"}, want: true},
		{name: "2", args: args{instring: "哈"}, want: false},
		{name: "3", args: args{instring: ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFirstLetterIsEngLower(tt.args.instring); got != tt.want {
				t.Errorf("isFirstLetterIsEngLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDialogueMerger(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"FixTimeline", "org"}, 4, false)

	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())
	//bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(filepath.Join(testRootDir, "2line-The Card Counter (2021) WEBDL-1080p.chinese(inside).ass"))
	bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(filepath.Join(testRootDir, "2line-英_1_0-3-35#150_36#360000.srt"))
	if err != nil {
		t.Fatal(err)
	}
	if bFind == false {
		t.Fatal("not find")
	}

	merger := NewDialogueMerger()
	for _, ex := range infoBase.DialoguesFilterEx {
		merger.Add(ex)
	}
	newEx := merger.Get()
	t.Logf("\n\n newEx: %d", len(newEx))
}
