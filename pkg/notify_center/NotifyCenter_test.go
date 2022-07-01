package notify_center

import (
	"testing"
)

func TestNewNotifyCenter(t *testing.T) {

	center := NewNotifyCenter("https://www.baidu.com/")
	center.Add("groupName", "Info asd 哈哈")
	center.Send()
}
