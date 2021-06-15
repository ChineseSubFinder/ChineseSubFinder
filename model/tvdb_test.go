package model

import (
	"testing"
)

func TestNewTVDB(t *testing.T) {

	configViper, err := InitConfigure()
	if err != nil {
		t.Fatal(err)
	}
	c, err := ReadConfig(configViper)
	if err != nil {
		t.Fatal(err)
	}

	tv := NewTVDB(c.TVdbApiKey)
	err = tv.login()
	if err != nil {
		t.Fatal(err)
	}
}
