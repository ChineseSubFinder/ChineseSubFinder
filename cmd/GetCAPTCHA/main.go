package main

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
)

func main() {

	err := Process()
	if err != nil {
		println(err.Error())
		return
	}
}

func Process() error {

	fmt.Println("-----------------------------------------")

	codeB64, err := backend.GetCode(config.GetConfig().DesURL)
	if err != nil {
		return err
	}

	err = backend.GitProcess(*config.GetConfig(), codeB64)
	if err != nil {
		return err
	}

	return nil
}
