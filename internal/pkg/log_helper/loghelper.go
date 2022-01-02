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
	"runtime"
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
	nowLoggerDir := getLoggerDir()
	if nowLoggerDir == "" {
		panic("Can't find logger dir by getLoggerDir()")
	}
	fileAbsPath := filepath.Join(nowLoggerDir, appName+".log")
	// 下面配置日志每隔 X 分钟轮转一个新文件，保留最近 X 分钟的日志文件，多余的自动清理掉。
	// create YYYYMMDDhhmmss for log
	timeStampString := time.Now().Format("20060102150405")
	writer, _ := rotatelogs.New(
		filepath.Join(nowLoggerDir, appName + "-" + timeStampString + ".log"),
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

func getLoggerDir() string {
	nowLoggerDir := ""
	sysType := runtime.GOOS
	if sysType == "linux" {
		nowLoggerDir = loggerDirLinux
	}
	if sysType == "windows" {
		nowLoggerDir = loggerDirWindows
	}
	if sysType == "darwin" {
		home, _ := os.UserHomeDir()
		nowLoggerDir = home + "/.config/chinesesubfinder/Logs" + loggerDirDarwin
	}
	return nowLoggerDir
}

var (
	logger  *logrus.Logger
	logOnce sync.Once
)

const (
	loggerDirLinux   = "/app/Logs/"
	loggerDirWindows = ""
	loggerDirDarwin  = ""
)
