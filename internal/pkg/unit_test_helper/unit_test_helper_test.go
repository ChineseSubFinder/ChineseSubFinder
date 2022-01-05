package unit_test_helper

import (
	"path/filepath"
	"testing"
)

func TestGetTestDataResourceRootPath(t *testing.T) {
	type args struct {
		resourceFolderNames []string
		goBackTimes         int
		useCopyData         bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "loghelper", args: args{
				resourceFolderNames: []string{"log_helper"},
				goBackTimes:         1,
				useCopyData:         false,
			},
			want: filepath.FromSlash("../ChineseSubFinder-TestData/log_helper"),
		},
		{
			name: "language", args: args{
				resourceFolderNames: []string{"language", "test"},
				goBackTimes:         1,
				useCopyData:         false,
			},
			want: filepath.FromSlash("../ChineseSubFinder-TestData/language/test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTestDataResourceRootPath(tt.args.resourceFolderNames, tt.args.goBackTimes, tt.args.useCopyData); got != tt.want {
				t.Errorf("GetTestDataResourceRootPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
