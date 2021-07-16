package pkg

import (
	"testing"
)

func TestNewNotifyCenter(t *testing.T) {

	configViper, err := InitConfigure()
	if err != nil {
		t.Fatal(err)
		return
	}
	config, err := ReadConfig(configViper)
	if err != nil {
		t.Fatal(err)
		return
	}
	center := NewNotifyCenter(config.WhenSubSupplierInvalidWebHook)
	center.Add("groupName", "Info asd 哈哈")
	center.Send()
}
