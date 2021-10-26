package log_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/config"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func NewLogHelper(appName string, level logrus.Level, maxAge time.Duration, rotationTime time.Duration) *logrus.Logger {

	Logger := &logrus.Logger{
		// Out:   os.Stderr,
		// Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}
	nowpath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pathRoot := filepath.Join(nowpath, "Logs")
	fileAbsPath := filepath.Join(pathRoot, appName+".log")
	// 下面配置日志每隔 X 分钟轮转一个新文件，保留最近 X 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		filepath.Join(pathRoot, appName+"--%Y%m%d%H%M--.log"),
		rotatelogs.WithLinkName(fileAbsPath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)

	Logger.SetLevel(level)
	Logger.SetOutput(io.MultiWriter(os.Stderr, writer))

	return Logger
}

func GetLogger() *logrus.Logger {
	logOnce.Do(func() {

		var level logrus.Level
		if config.GetConfig().DebugMode == true {
			level = logrus.DebugLevel
		} else {
			level = logrus.InfoLevel
		}

		logger = NewLogHelper("ChineseSubFinder", level, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
	})
	return logger
}

var logger *logrus.Logger
var logOnce sync.Once
