package pre_job

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/hot_fix"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/sirupsen/logrus"
)

type PreJob struct {
	stageName string
	gError    error

	sets *settings.Settings
	log  *logrus.Logger
}

func NewPreJob(sets *settings.Settings, log *logrus.Logger) *PreJob {
	return &PreJob{sets: sets, log: log}
}

func (p *PreJob) HotFix() *PreJob {

	if p.gError != nil {
		p.log.Infoln("Skip PreJob.Check()")
		return p
	}
	p.stageName = stageNameCHotFix

	defer func() {
		p.log.Infoln("PreJob.HotFix() End")
	}()
	p.log.Infoln("PreJob.HotFix() Start...")
	// ------------------------------------------------------------------------
	// 开始修复
	p.log.Infoln(common2.NotifyStringTellUserWait)
	err := hot_fix.HotFixProcess(types.HotFixParam{
		MovieRootDirs:  p.sets.CommonSettings.MoviePaths,
		SeriesRootDirs: p.sets.CommonSettings.SeriesPaths,
	})
	if err != nil {
		p.log.Errorln("hot_fix.HotFixProcess()", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreJob) ChangeSubNameFormat() *PreJob {

	if p.gError != nil {
		p.log.Infoln("Skip PreJob.ChangeSubNameFormat()")
		return p
	}
	p.stageName = stageNameChangeSubNameFormat
	defer func() {
		p.log.Infoln("PreJob.ChangeSubNameFormat() End")
	}()
	p.log.Infoln("PreJob.ChangeSubNameFormat() Start...")
	// ------------------------------------------------------------------------
	/*
		字幕命名格式转换，需要数据库支持
		如果数据库没有记录经过转换，那么默认从 Emby 的格式作为检测的起点，转换到目标的格式
		然后需要在数据库中记录本次的转换结果
	*/
	p.log.Infoln(common2.NotifyStringTellUserWait)
	renameResults, err := sub_formatter.SubFormatChangerProcess(
		p.sets.CommonSettings.MoviePaths,
		p.sets.CommonSettings.SeriesPaths,
		common.FormatterName(p.sets.AdvancedSettings.SubNameFormatter))
	// 出错的文件有哪一些
	for s, i := range renameResults.ErrFiles {
		p.log.Errorln("reformat ErrFile:"+s, i)
	}
	if err != nil {
		p.log.Errorln("SubFormatChangerProcess() Error", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreJob) ReloadBrowser() *PreJob {

	if p.gError != nil {
		p.log.Infoln("Skip PreJob.ReloadBrowser()")
		return p
	}
	p.stageName = stageNameReloadBrowser
	defer func() {
		p.log.Infoln("PreJob.ReloadBrowser() End")
	}()
	p.log.Infoln("PreJob.ReloadBrowser() Start...")
	// ------------------------------------------------------------------------
	// ReloadBrowser 提前把浏览器下载好
	rod_helper.ReloadBrowser()
	return p
}

func (p *PreJob) Wait() error {
	defer func() {
		p.log.Infoln("PreJob.Wait() Done.")
	}()
	if p.gError != nil {
		outErrString := "PreJob.Wait() Get Error, " + "stageName:" + p.stageName + " -- " + p.gError.Error()
		p.log.Errorln(outErrString)
		return errors.New(outErrString)
	} else {
		return nil
	}
}

const (
	stageNameCHotFix             = "HotFix"
	stageNameChangeSubNameFormat = "ChangeSubNameFormat"
	stageNameReloadBrowser       = "ReloadBrowser"
)
