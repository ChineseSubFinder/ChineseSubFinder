package backend

import (
	b64 "encoding/base64"
	"errors"
	"github.com/allanpk716/rod_helper"
	"regexp"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

func GetCode(log *logrus.Logger, codeUrl string) (string, error) {

	defer func() {
		log.Infoln("End Get Code")
	}()

	log.Infoln("Start Get Code...")
	browser, err := rod_helper.NewBrowserBase("", "", "", false, true)
	if err != nil {
		return "", err
	}
	defer func() {
		browser.Close()
	}()
	page, err := browser.Browser.Page(proto.TargetCreateTarget{URL: codeUrl})
	if err != nil {
		return "", err
	}
	defer func() {
		_ = page.Close()
	}()
	err = page.WaitLoad()
	if err != nil {
		return "", err
	}
	htmlString, err := page.HTML()
	if err != nil {
		return "", err
	}

	code := ""
	var re = regexp.MustCompile(`(?m)>\d{6}<`)
	parts := re.FindAllString(htmlString, -1)
	if parts == nil || len(parts) <= 0 {
		return "", errors.New("code not found")
	} else {
		code = strings.ReplaceAll(parts[0], "<", "")
		code = strings.ReplaceAll(code, ">", "")
	}

	sEnc := b64.StdEncoding.EncodeToString([]byte(code))

	log.Infoln("code:", code)
	log.Infoln("sEnc:", sEnc)

	return sEnc, nil
}
