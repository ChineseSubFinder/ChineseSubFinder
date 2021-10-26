package notify_center

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/config"
	"testing"
)

func TestNewNotifyCenter(t *testing.T) {

	config := config.GetConfig()
	center := NewNotifyCenter(config.WhenSubSupplierInvalidWebHook)
	center.Add("groupName", "Info asd 哈哈")
	center.Send()
}
