package my_util

import (
	"os"
	"strconv"
	"strings"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/sirupsen/logrus"
)

func ReadCustomPortFile(log *logrus.Logger) int {
	if IsFile(customPort) == false {
		return defPort
	} else {
		bytes, err := os.ReadFile(customPort)
		if err != nil {
			log.Errorln("ReadFile CustomPort Error", err)
			log.Infoln("Use DefPort", defPort)
			return defPort
		}

		atoi, err := strconv.Atoi(string(bytes))
		if err != nil {
			log.Errorln("Atoi CustomPort Error", err)
			log.Infoln("Use DefPort", defPort)
			return defPort
		}

		log.Infoln("Use CustomPort", atoi)
		return atoi
	}
}

func ReadCustomAuthFile(log *logrus.Logger) bool {
	if IsFile(customAuth) == false {
		return false
	} else {
		bytes, err := os.ReadFile(customAuth)
		if err != nil {
			log.Errorln("ReadFile CustomAuth Error", err)
			return false
		}

		nowContent := string(bytes)
		authStings := strings.Split(nowContent, "@@@@")
		if len(authStings) != 3 {
			log.Errorln("ReadFile CustomAuth Error", err)
			return false
		}

		global_value.SetBaseKey(authStings[0])
		global_value.SetAESKey16(authStings[1])
		global_value.SetAESIv16(authStings[2])

		log.Infoln("Use CustomAuth")
		return true
	}
}

const (
	defPort    = 19035
	customPort = "CustomPort"
	customAuth = "CustomAuth"
)
