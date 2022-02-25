package log_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/regex_things"
	"github.com/allanpk716/ChineseSubFinder/internal/types/log_hub"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
	这里独立出来一个按扫描次为单位的日志模块，暂时没有打算替换原有的日志模块
	仅仅是为了更好地获取、展示日志，也方便后续的问题反馈提交
	考虑到这里，暂定这里的日志都需要存储到硬盘中，且有两种方式获取：
		1. 从 http 接口查询完成的多次的日志
		2. 从 ws 接口中获取当前正在进行扫描的日志
	既然是没有替换的打算，那么就使用 logrus 的 hook 接口去完成额外日志的记录即可，也就是在“每次”扫描的开始和结束进行标记，然后拆分成多次的日志好了
*/

func init() {
	// 第一次运行需要清理、读取一次
	cleanAndLoadOnceLogs()
}

type LoggerHub struct {
	onceLogger *logrus.Logger // 一次扫描日志的实例
	onceStart  bool
}

func NewLoggerHub() *LoggerHub {
	return &LoggerHub{}
}

func (lh *LoggerHub) Levels() []logrus.Level {
	// 记录全级别
	return []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (lh *LoggerHub) Fire(entry *logrus.Entry) error {

	if entry.Message == OnceSubsScanStart {
		// 收到日志的标志位，需要新开一个
		if lh.onceStart == false {
			lh.onceLogger = newOnceLogger()
			lh.onceStart = true
		}
		return nil
	} else if entry.Message == OnceSubsScanEnd {
		// “一次”扫描的结束标志位
		lh.onceStart = false
		if onceLoggerFile != nil {
			_ = onceLoggerFile.Close()
			onceLoggerFile = nil
		}
		return nil
	}

	if lh.onceStart == false {
		// 如果没有发现开启一次扫描的记录标志位，那么就不进行日志的写入
		return nil
	}

	switch entry.Level {
	case logrus.TraceLevel:
		lh.onceLogger.Traceln(entry.Message)
	case logrus.DebugLevel:
		lh.onceLogger.Debugln(entry.Message)
	case logrus.InfoLevel:
		lh.onceLogger.Infoln(entry.Message)
	case logrus.WarnLevel:
		lh.onceLogger.Warningln(entry.Message)
	case logrus.ErrorLevel:
		lh.onceLogger.Errorln(entry.Message)
	case logrus.FatalLevel:
		lh.onceLogger.Fatalln(entry.Message)
	case logrus.PanicLevel:
		lh.onceLogger.Panicln(entry.Message)
	}

	return nil
}

// GetRecentOnceLogs 获取最近多少次扫描的日志信息
func GetRecentOnceLogs(getHowMany int) []log_hub.OnceLog {
	defer func() {
		onceLogsLock.Unlock()
	}()
	onceLogsLock.Lock()

	tmpOnceLogs := make([]log_hub.OnceLog, 0)

	nowGetCount := getHowMany
	if nowGetCount > len(onceLogs) {
		nowGetCount = len(onceLogs)
	}
	for i := 0; i < nowGetCount; i++ {

		tmpOnceLogs = append(tmpOnceLogs, onceLogs[i])
	}

	return tmpOnceLogs
}

func newOnceLogger() *logrus.Logger {

	var err error
	Logger := logrus.New()
	Logger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	}
	nowTime := time.Now()
	pathRoot := filepath.Join(global_value.ConfigRootDirFPath, "Logs")
	fileName := fmt.Sprintf(onceLogPrefix+"%v.log", nowTime.Unix())
	fileAbsPath := filepath.Join(pathRoot, fileName)
	if onceLoggerFile != nil {
		_ = onceLoggerFile.Close()
		onceLoggerFile = nil
	}
	// 注意这个函数的调用时机
	cleanAndLoadOnceLogs()

	onceLoggerFile, err = os.OpenFile(fileAbsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		GetLogger().Panicln("newOnceLogger.OpenFile", err)
	}
	Logger.SetOutput(onceLoggerFile)

	return Logger
}

// cleanAndLoadOnceLogs 调用的时机，一定是要在新开一个日志前，且把上一个日志的文件流关闭的时候
func cleanAndLoadOnceLogs() {
	defer func() {
		onceLogsLock.Unlock()
	}()

	onceLogsLock.Lock()

	onceLogs = make([]log_hub.OnceLog, 0)

	pathRoot := filepath.Join(global_value.ConfigRootDirFPath, "Logs")
	// 扫描当前日志存储目录下有多少个符合要求的 Once- 日志
	// 确保有且仅有最近的 20 次扫描日志记录存在即可
	matches, err := filepath.Glob(filepath.Join(pathRoot, onceLogPrefix+"*.log"))
	if err != nil {
		GetLogger().Panicln("cleanAndLoadOnceLogs.Glob", err)
	}
	if len(matches) > onceLogMaxCount {
		// 需要清理多余的
		// 保存的文件名是 Once-unixTime.log 做为前提
		// 这里假定查询出来的都是正序排序
		for i := 0; i <= len(matches)-1-onceLogMaxCount; i++ {

			_, err := os.Stat(matches[i])
			if err != nil {
				continue
			}
			_ = os.Remove(matches[i])
		}
		// 将有存在价值的“单次”日志缓存到内存中，供 Web API 查询
		matches, err = filepath.Glob(filepath.Join(pathRoot, onceLogPrefix+"*.log"))
		if err != nil {
			GetLogger().Panicln("cleanAndLoadOnceLogs.Glob", err)
		}
	}
	j := 0
	for i := len(matches) - 1; i >= 0; i-- {
		// 需要逆序放入到缓存中，因为定义的是 Index = 0 是最新的一个完成的扫描日志
		err = readLogFile(j, matches[i])
		if err != nil {
			j++
			continue
		}
		j++
	}
}

func readLogFile(index int, filePath string) error {

	fBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	matched := regex_things.ReMathLogOneLine.FindAllStringSubmatch(string(fBytes), -1)
	if matched == nil || len(matched) < 1 {
		GetLogger().Debugln("readLogFile can't found ReMathLogOneLine, Skip")
		return nil
	}

	nowOnceLog := log_hub.NewOnceLog(index)
	for _, oneLine := range matched {
		nowOnceLog.LogLines = append(nowOnceLog.LogLines,
			*log_hub.NewOneLine(
				oneLine[2], // Level
				oneLine[4], // DateTime
				oneLine[5], // Content
			))
	}

	onceLogs = append(onceLogs, *nowOnceLog)

	return nil
}

var (
	onceLoggerFile *os.File
	onceLogs       = make([]log_hub.OnceLog, 0)
	onceLogsLock   sync.Mutex
)

const (
	onceLogMaxCount   = 20
	onceLogPrefix     = "Once-"
	OnceSubsScanStart = "OneTimeSubtitleScanStart"
	OnceSubsScanEnd   = "OneTimeSubtitleScanEnd"
)
