package log_helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/log_hub"

	"github.com/huandu/go-clone"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

/*
	这里独立出来一个按扫描次为单位的日志模块，暂时没有打算替换原有的日志模块
	仅仅是为了更好地获取、展示日志，也方便后续的问题反馈提交
	考虑到这里，暂定这里的日志都需要存储到硬盘中，且有两种方式获取：
		1. 从 http 接口查询完成的多次的日志
		2. 从 ws 接口中获取当前正在进行扫描的日志
	既然是没有替换的打算，那么就使用 logrus 的 hook 接口去完成额外日志的记录即可，也就是在“每次”扫描的开始和结束进行标记，然后拆分成多次的日志好了
*/

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

	// 如果是一次扫描的开始
	if strings.HasPrefix(entry.Message, OnceSubsScanStart) == true {
		// 收到日志的标志位，需要新开一个
		if lh.onceStart == false {
			// 这个日志的前缀是 OnceSubsScanStart ，然后通过 # 进行分割，得到任务的 ID

			names := strings.Split(entry.Message, "#")
			if len(names) > 1 {
				lh.onceLogger = newOnceLogger(names[1])
			} else {
				lh.onceLogger = newOnceLogger(fmt.Sprintf("%v", time.Now().Unix()))
			}
			lh.onceStart = true
			// 既然新的一次开始，就实例化新的实例出来使用
			onceLogsLock.Lock()
			onceLog4Running = log_hub.NewOnceLog(0)
			onceLogsLock.Unlock()
		}
		return nil
	} else if entry.Message == OnceSubsScanEnd {
		// “一次”扫描的结束标志位
		lh.onceStart = false

		// 注意这个函数的调用时机
		CleanAndLoadOnceLogs()

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

	onceLogsLock.Lock()
	onceLog4Running.LogLines = append(onceLog4Running.LogLines, *log_hub.NewOneLine(
		entry.Level.String(),
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Message))
	onceLogsLock.Unlock()

	return nil
}

// GetOnceLog4Running 当前正在运行任务的日志
func GetOnceLog4Running() *log_hub.OnceLog {

	var nowOnceRunningLog *log_hub.OnceLog

	onceLogsLock.Lock()
	nowOnceRunningLog = clone.Clone(onceLog4Running).(*log_hub.OnceLog)
	onceLogsLock.Unlock()

	return nowOnceRunningLog
}

// GetSpiltOnceLog 拆分到一行一个，没有锁，所以需要考虑并发问题
func GetSpiltOnceLog(log *log_hub.OnceLog) []*log_hub.OnceLog {

	if log == nil {
		return nil
	}

	var outList = make([]*log_hub.OnceLog, len(log.LogLines))
	for i := 0; i < len(log.LogLines); i++ {
		outList[i] = &log_hub.OnceLog{
			LogLines: []log_hub.OneLine{log.LogLines[i]},
		}
	}

	return outList
}

func newOnceLogger(logFileName string) *logrus.Logger {

	var err error
	Logger := logrus.New()
	Logger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	}
	pathRoot := filepath.Join(pkg.ConfigRootDirFPath(), "Logs")
	fileName := fmt.Sprintf(common.OnceLogPrefix+"%v.log", logFileName)
	fileAbsPath := filepath.Join(pathRoot, fileName)

	// 注意这个函数的调用时机
	CleanAndLoadOnceLogs()

	onceLoggerFile, err = os.OpenFile(fileAbsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}
	Logger.SetOutput(onceLoggerFile)

	return Logger
}

// CleanAndLoadOnceLogs 调用的时机，一定是要在新开一个日志前，且把上一个日志的文件流关闭的时候
func CleanAndLoadOnceLogs() {
	defer func() {
		onceLogsLock.Unlock()
	}()

	if onceLoggerFile != nil {
		_ = onceLoggerFile.Close()
		onceLoggerFile = nil
	}

	onceLogsLock.Lock()

	pathRoot := filepath.Join(pkg.ConfigRootDirFPath(), "Logs")
	// 扫描当前日志存储目录下有多少个符合要求的 Once- 日志
	// 确保有且仅有最近的 20 次扫描日志记录存在即可
	matches, err := filepath.Glob(filepath.Join(pathRoot, common.OnceLogPrefix+"*.log"))
	if err != nil {
		panic(err)
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
		matches, err = filepath.Glob(filepath.Join(pathRoot, common.OnceLogPrefix+"*.log"))
		if err != nil {
			panic(err)
		}
	}
}

var (
	onceLoggerFile  *os.File                // 单次扫描保存 Log 文件的实例
	onceLogsLock    sync.Mutex              // 对应的锁
	onceLog4Running = log_hub.NewOnceLog(0) // 当前正在扫描时候日志的日志内容实例，注意，开启任务不代表就在扫描
)

const (
	onceLogMaxCount = 10000

	OnceSubsScanStart = "OneTimeSubtitleScanStart"
	OnceSubsScanEnd   = "OneTimeSubtitleScanEnd"
)
