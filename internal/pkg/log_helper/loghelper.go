package log_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
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

	Logger := logrus.New()
	Logger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	}
	pathRoot := filepath.Join(global_value.ConfigRootDirFPath(), "Logs")
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

	// 可以输出函数调用还文件位置
	//if level == logrus.DebugLevel {
	//	Logger.SetReportCaller(true)
	//}
	return Logger
}

// SetLoggerName 如果是 ChineseSubFinder 调用则无需使用，其他子程序用的时候，为了区分日子名称，需要设置
func SetLoggerName(logName string) {
	if logName == "" {
		panic("Need Set Logger Name")
	}
	logNameBase = logName
}

func GetLogger(reload ...bool) *logrus.Logger {

	if len(reload) > 0 {
		if reload[0] == true {
			logInit()
		}
	}
	oneBase.Do(logInit)

	return loggerBase
}

func logInit() {
	var level logrus.Level
	// 之前是读取配置文件，现在改为，读取当前目录下，是否有一个特殊的文件，有则启动 Debug 日志级别
	// 那么怎么写入这个文件，就靠额外的逻辑控制了
	if isFile(filepath.Join(global_value.ConfigRootDirFPath(), DebugFileName)) == true {
		level = logrus.DebugLevel
	} else {
		level = logrus.InfoLevel
	}

	if logNameBase == "" {
		// 默认不设置的时候就是这个
		logNameBase = LogNameChineseSubFinder
	}

	loggerBase = NewLogHelper(logNameBase, level, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)

	loggerBase.AddHook(NewLoggerHub())
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
	if isFile(filepath.Join(global_value.ConfigRootDirFPath(), DebugFileName)) == true {
		return nil
	}
	f, err := os.Create(filepath.Join(global_value.ConfigRootDirFPath(), DebugFileName))
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

	if isFile(filepath.Join(global_value.ConfigRootDirFPath(), DebugFileName)) == false {
		return nil
	}
	err := os.Remove(filepath.Join(global_value.ConfigRootDirFPath(), DebugFileName))
	if err != nil {
		return err
	}
	return nil
}

const DebugFileName = "opendebuglog"

const (
	LogNameChineseSubFinder = "ChineseSubFinder"
	LogNameGetCAPTCHA       = "GetCAPTCHA"
	LogNameBackEnd          = "BackEnd"
)

var (
	logNameBase = ""
	oneBase     = sync.Once{}
	loggerBase  *logrus.Logger
)
