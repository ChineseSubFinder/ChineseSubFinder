package log_helper

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/extend_log"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"

	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func NewLogHelper(appName string, logStorePath string, level logrus.Level, maxAge time.Duration, rotationTime time.Duration, extendLog ...settings.ExtendLog) *logrus.Logger {

	Logger := logrus.New()
	Logger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	}
	pathRoot := filepath.Join(logStorePath, "Logs")
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

	if len(extendLog) > 0 {
		exLog := extend_log.ExtendLogEx{}
		exLog.AddHook(Logger, extendLog[0])
	}
	// 可以输出函数调用还文件位置
	//if level == logrus.DebugLevel {
	//	Logger.SetReportCaller(true)
	//}
	return Logger
}

func isFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// WriteDebugFile 写入开启 Debug 级别日志记录的特殊文件，注意这个最好是在主程序中调用，这样就跟主程序在一个目录下生成，log 去检测是否存在才有意义
func WriteDebugFile() error {
	if isFile(filepath.Join(pkg.ConfigRootDirFPath(), DebugFileName)) == true {
		return nil
	}
	f, err := os.Create(filepath.Join(pkg.ConfigRootDirFPath(), DebugFileName))
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	return nil
}

// DeleteDebugFile 删除开启 Debug 级别日志记录的特殊文件
func DeleteDebugFile() error {

	if isFile(filepath.Join(pkg.ConfigRootDirFPath(), DebugFileName)) == false {
		return nil
	}
	err := os.Remove(filepath.Join(pkg.ConfigRootDirFPath(), DebugFileName))
	if err != nil {
		return err
	}
	return nil
}

func GetLogger4Tester() *logrus.Logger {
	if logger4Tester == nil {
		logger4Tester = NewLogHelper(LogNameChineseSubFinder, os.TempDir(), logrus.DebugLevel, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)

	}
	return logger4Tester
}

const DebugFileName = "opendebuglog"

const (
	LogNameChineseSubFinder    = "ChineseSubFinder"
	LogNameGetCAPTCHA          = "GetCAPTCHA"
	LogNameBackEnd             = "BackEnd"
	LogNameCliSubTimelineFixer = "SubTimelineFixer"
)

var (
	logger4Tester *logrus.Logger
)
