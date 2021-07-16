package notify_center

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"testing"
)

func TestNewNotifyCenter(t *testing.T) {

	config := pkg.GetConfig()
	center := NewNotifyCenter(config.WhenSubSupplierInvalidWebHook)
	center.Add("groupName", "Info asd 哈哈")
	center.Send()
}
