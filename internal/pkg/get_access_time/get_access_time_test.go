package get_access_time

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestGetAccessTime_GetAccessTime(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_parser", "org"}, 4, false)
	fileFPath := filepath.Join(testRootDir, "[xunlei]_0_C3A5CUsers5CAdministrator5CDesktop5CThe Boss Baby Family Business_S0E0.ass")

	g := GetAccessTimeEx{}
	println(g.GetOSName())
	accessTime, err := g.GetAccessTime(fileFPath)
	if err != nil {
		t.Fatal(err)
	}
	println(accessTime.String())
}
