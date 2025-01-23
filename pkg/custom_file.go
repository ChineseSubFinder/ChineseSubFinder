package pkg

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func ReadCustomHostFile(log *logrus.Logger) string {
	if IsFile(customHost) == false {
		return defHost
	} else {
		bytes, err := os.ReadFile(customHost)
		if err != nil {
			log.Errorln("ReadFile customHost Error", err)
			log.Infoln("Use DerHost '0.0.0.0'")
			return defHost
		}
		nowContent := string(bytes)
		host := net.ParseIP(nowContent)
		if host == nil || host.To4() == nil {
			log.Errorln("ParseIP customHost (Invalid IPv4) Error", err)
			log.Infoln("Use DerHost '0.0.0.0'")
			return defHost
		} else {
			log.Infoln("Use CustomHost", nowContent)
			return nowContent
		}
	}
}

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

		SetBaseKey(authStings[0])
		SetAESKey16(authStings[1])
		SetAESIv16(authStings[2])

		log.Infoln("Use CustomAuth")
		return true
	}
}

const (
	defPort    = 19035
	defHost    = "0.0.0.0"
	customPort = "CustomPort"
	customHost = "CustomHost"
	customAuth = "CustomAuth"
)
