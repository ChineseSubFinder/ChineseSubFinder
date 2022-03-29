package backend

import (
	b64 "encoding/base64"
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/go-rod/rod/lib/proto"
	"regexp"
	"strings"
)

func GetCode(codeUrl string) (string, error) {

	defer func() {
		log_helper.GetLogger().Infoln("End Get Code")
	}()

	log_helper.GetLogger().Infoln("Start Get Code...")
	browser, err := rod_helper.NewBrowser("", false)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = browser.Close()
	}()
	page, err := browser.Page(proto.TargetCreateTarget{URL: codeUrl})
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

	log_helper.GetLogger().Infoln("code:", code)
	log_helper.GetLogger().Infoln("sEnc:", sEnc)

	return sEnc, nil
}
