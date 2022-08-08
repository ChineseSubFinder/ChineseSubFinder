package rod_helper

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

func TestNewMultiBrowser(t *testing.T) {

	browserOptions := NewBrowserOptions(log_helper.GetLogger4Tester(), true, settings.GetSettings())
	browserOptions.SetXrayPoolUrl("127.0.0.1")
	browserOptions.SetXrayPoolPort("19035")
	b := NewMultiBrowser(browserOptions)
	println(b)
}
