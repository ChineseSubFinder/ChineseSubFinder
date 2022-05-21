package main

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)


/**
字幕时间轴修复命令行
使用方法：
go run main.go -vp ${videoPath} -sp ${subtitlePath}
${videPath} -> 视频文件路径，需要指定对应的视频文件
${subtitlePath} -> 字幕文件路径，需要指定对应的字幕文件

逻辑:
1. 执行 SubTimelineFixerHelperEx 检查 - 确认已经安装了ffmpeg 和 ffprobe
2. 执行 SubTimelineFixerHelperEx 的 process操作

 */

var loggerBase *logrus.Logger

func newLog() *logrus.Logger {
	logger := log_helper.NewLogHelper(log_helper.LogNameCliSubTimelineFixer,
		global_value.ConfigRootDirFPath(),
		logrus.InfoLevel, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
	return logger
}

func main() {

	var videoPath string
	var subtitlesPath string

	loggerBase = newLog()

	app := &cli.App{
		Name:  "Subtitle Timeline Fixer",
		Usage: "Fix the subtitle timeline according to the video",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "videoPath",
				Aliases: []string{"vp"},
				Usage: "Specify `video file path`",
				Destination: &videoPath,
				Required: true,
			},

			&cli.StringFlag{
				Name:  "subtitlesPath",
				Aliases: []string{"sp"},
				Usage: "Specify `subtitles file path`",
				Destination: &subtitlesPath,
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			videoPath = strings.TrimSpace(videoPath)
			subtitlesPath = strings.TrimSpace(subtitlesPath)
			if videoPath != ""  && subtitlesPath != "" {
				var fixerSetting = settings.NewTimelineFixerSettings()
				var subTimelineFixerHelper = sub_timeline_fixer.NewSubTimelineFixerHelperEx(loggerBase, *fixerSetting);
				if subTimelineFixerHelper.Check() {
					subTimelineFixerHelper.Process(videoPath, subtitlesPath);
				} else {
					println("check subtitles timeline fixer helper failed.")
				}
			} else {
				println("need provide video path (-vp) and subtitle path (-sp)")
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

