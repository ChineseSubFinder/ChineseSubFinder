package pre_job

import (
	"errors"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"
	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/hot_fix"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
	"github.com/sirupsen/logrus"
)

type PreJob struct {
	stageName     string
	gError        error
	isDone        bool
	log           *logrus.Logger
	renameResults sub_formatter.RenameResults
}

func NewPreJob(log *logrus.Logger) *PreJob {
	return &PreJob{log: log, isDone: false}
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
	err := hot_fix.HotFixProcess(p.log, types.HotFixParam{
		MovieRootDirs:  settings.Get().CommonSettings.MoviePaths,
		SeriesRootDirs: settings.Get().CommonSettings.SeriesPaths,
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
	var err error
	p.renameResults, err = sub_formatter.SubFormatChangerProcess(p.log,
		settings.Get().CommonSettings.MoviePaths,
		settings.Get().CommonSettings.SeriesPaths,
		common.FormatterName(settings.Get().AdvancedSettings.SubNameFormatter))
	// 出错的文件有哪一些
	for s, i := range p.renameResults.ErrFiles {
		p.log.Errorln("reformat ErrFile:"+s, i)
	}
	if err != nil {
		p.log.Errorln("SubFormatChangerProcess() Error", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreJob) Wait() error {
	defer func() {
		p.isDone = true
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

func (p *PreJob) GetStageName() string {
	return p.stageName
}

func (p *PreJob) IsDone() bool {
	return p.isDone
}

func (p *PreJob) GetRenameResults() sub_formatter.RenameResults {
	return p.renameResults
}

func (p *PreJob) GetGError() error {
	return p.gError
}

const (
	stageNameCHotFix             = "HotFix"
	stageNameChangeSubNameFormat = "ChangeSubNameFormat"
	stageNameReloadBrowser       = "ReloadBrowser"
)
