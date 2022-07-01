package sub_timeline_fixer

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"os"
)

func SaveStaticLineV1(saveFPath string, infoBaseName, infoSrcName string,
	per, oldMean, OldSd, NewMean, NewSd float64, xAxis []string,
	startDiffTimeLineData, endDiffTimeLineData []opts.LineData) error {
	// 1.New 一个条形图对象
	bar := charts.NewLine()
	// 2.设置 标题 和 子标题
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    infoBaseName + " <-->" + infoSrcName,
		Subtitle: fmt.Sprintf("One Dialogue Start-End(Blue and Green) Base -> Src Start (newLen / orgLen = %f) \r\nOldMean: %f OldSD: %f -- NewMean: %f NewSD: %f", per, oldMean, OldSd, NewMean, NewSd),
	}))

	// 3.设置 数据组
	bar.SetXAxis(xAxis).
		AddSeries("Start Time Diff", startDiffTimeLineData).
		AddSeries("End Time Diff", endDiffTimeLineData)
	// 4.绘图 生成html
	f, err := os.Create(saveFPath)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	err = bar.Render(f)
	if err != nil {
		return err
	}

	return nil
}

func SaveStaticLineV2(name, saveFPath string, xAxis []string, timeLineOrgData []opts.LineData) error {

	// 1.New 一个条形图对象
	bar := charts.NewLine()
	// 2.设置 标题 和 子标题
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: name + " VAD",
	}))
	// 3.设置 数据组
	bar.SetXAxis(xAxis).
		AddSeries(name+" VAD", timeLineOrgData)
	// 4.绘图 生成html
	outfile, err := os.Create(saveFPath)
	defer func() {
		_ = outfile.Close()
	}()
	if err != nil {
		return err
	}
	err = bar.Render(outfile)
	if err != nil {
		return err
	}
	return nil
}

func SaveStaticLineV3(name, saveFPath string, xAxis []string, timeLineOrgData, fftData []opts.LineData) error {

	// 1.New 一个条形图对象
	bar := charts.NewLine()
	// 2.设置 标题 和 子标题
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: name + " VAD",
	}))
	// 3.设置 数据组
	bar.SetXAxis(xAxis).
		AddSeries(name+" VAD", timeLineOrgData) //.
		//AddSeries(name+" FFT", fftData)
	// 4.绘图 生成html
	outfile, err := os.Create(saveFPath)
	defer func() {
		_ = outfile.Close()
	}()
	if err != nil {
		return err
	}
	err = bar.Render(outfile)
	if err != nil {
		return err
	}
	return nil
}
