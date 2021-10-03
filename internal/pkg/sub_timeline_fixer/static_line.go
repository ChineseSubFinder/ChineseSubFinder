package sub_timeline_fixer

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"os"
)

func SaveStaticLine(saveFPath string, infoBaseName, infoSrcName string,
	mean, sd float64, xAxis []string,
	startDiffTimeLineData, endDiffTimeLineData []opts.LineData) error {
	// 1.New 一个条形图对象
	bar := charts.NewLine()
	// 2.设置 标题 和 子标题
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    infoBaseName + " <-->" + infoSrcName,
		Subtitle: fmt.Sprintf("Base -> Src Mean: %f SD: %f", mean, sd),
	}))

	// 3.设置 数据组
	bar.SetXAxis(xAxis).
		AddSeries("Start Time Diff", startDiffTimeLineData).
		AddSeries("End Time Diff", endDiffTimeLineData)
	// 4.绘图 生成html
	f, _ := os.Create(saveFPath)
	err := bar.Render(f)
	if err != nil {
		return err
	}

	return nil
}
