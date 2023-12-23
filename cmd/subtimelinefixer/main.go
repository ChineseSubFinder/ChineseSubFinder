package main

import (
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/sirupsen/logrus"
)

/*
	字幕时间轴修复命令行
	使用方法：
	go run main.go -vp ${videoPath} -sp ${subtitlePath}
	${videPath} -> 视频文件路径，需要指定对应的视频文件
	${subtitlePath} -> 字幕文件路径，需要指定对应的字幕文件

	逻辑:
	1. 执行 SubTimelineFixerHelperEx 检查 - 确认已经安装了ffmpeg 和 ffprobe
	2. 执行 SubTimelineFixerHelperEx 的 process操作

	编译：
	通过`go build -o fixer`编译出可直接执行的文件。
*/
var loggerBase *logrus.Logger

func newLog() *logrus.Logger {
	logger := log_helper.NewLogHelper(log_helper.LogNameCliSubTimelineFixer,
		pkg.ConfigRootDirFPath(),
		logrus.InfoLevel, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
	return logger
}

func main() {

	//var videoPath string
	//var subtitlesPath string

	loggerBase = newLog()

	//videoPath = "X:\\连续剧\\All Creatures Great and Small (2020)\\Season 1\\All Creatures Great and Small (2020) - S01E04 - A Tricki Case WEBDL-1080p.mkv"
	////videoPath = "X:\\连续剧\\All Creatures Great and Small (2020)\\Season 1\\All Creatures Great and Small (2020) - S01E05 - All's Fair WEBDL-1080p.mkv"
	////subtitlesPath = "D:\\Subtitle\\123.ass"
	////subtitlesPath = "D:\\Subtitle\\123.srt"
	////subtitlesPath = "D:\\Subtitle\\英_2.srt"
	//subtitlesPath = "D:\\Subtitle\\All Creatures Great and Small (2020) - S01E04 - A Tricki Case WEBDL-1080p.chinese(简,manual).default.srt"
	//
	//var fixerSetting = settings.NewTimelineFixerSettings()
	//fixerSetting.ThreadCount = 1
	//fixerSetting.MinOffset = 3
	//var subTimelineFixerHelper = sub_timeline_fixer.NewSubTimelineFixerHelperEx(loggerBase, *fixerSetting)
	//if subTimelineFixerHelper.Check() {
	//	err := subTimelineFixerHelper.Process(videoPath, subtitlesPath, sub_timeline_fixer.ProcessTypeByAudioFile)
	//	if err != nil {
	//		println(err)
	//	}
	//} else {
	//	println("check subtitles timeline fixer helper failed.")
	//}
	//
	//println("Done.")

	//app := &cli.App{
	//	Name:  "Subtitle Timeline Fixer",
	//	Usage: "Fix the subtitle timeline according to the video",
	//	Flags: []cli.Flag{
	//		&cli.StringFlag{
	//			Name:        "videoPath",
	//			Aliases:     []string{"vp"},
	//			Usage:       "Specify `video file path`",
	//			Destination: &videoPath,
	//			Required:    true,
	//		},
	//
	//		&cli.StringFlag{
	//			Name:        "subtitlesPath",
	//			Aliases:     []string{"sp"},
	//			Usage:       "Specify `subtitles file path`",
	//			Destination: &subtitlesPath,
	//			Required:    true,
	//		},
	//	},
	//	Action: func(c *cli.Context) error {
	//		videoPath = strings.TrimSpace(videoPath)
	//		subtitlesPath = strings.TrimSpace(subtitlesPath)
	//		if videoPath != "" && subtitlesPath != "" {
	//			var fixerSetting = settings.NewTimelineFixerSettings()
	//			var subTimelineFixerHelper = sub_timeline_fixer.NewSubTimelineFixerHelperEx(loggerBase, *fixerSetting)
	//			if subTimelineFixerHelper.Check() {
	//				subTimelineFixerHelper.Process(videoPath, subtitlesPath)
	//			} else {
	//				println("check subtitles timeline fixer helper failed.")
	//			}
	//		} else {
	//			println("need provide video path (-vp) and subtitle path (-sp)")
	//		}
	//		return nil
	//	},
	//}
	//
	//err := app.Run(os.Args)
	//if err != nil {
	//	log.Fatal(err)
	//}
}
