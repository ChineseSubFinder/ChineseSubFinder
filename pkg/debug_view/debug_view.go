package debug_view

import (
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"os"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func SaveDebugChart(subUnit sub_helper.SubUnit, title, subTitle string) error {

	return SaveDebugChartBase(subUnit.VADList, title, subTitle)
}

func SaveDebugChartBase(vadList []vad.VADInfo, title, subTitle string) error {

	line := charts.NewBar()
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    title,
		Subtitle: subTitle,
	}))
	// 构建 X 轴
	xAxis := make([]string, len(vadList))
	for i := 0; i < len(vadList); i++ {
		xAxis[i] = fmt.Sprintf("%d", i)
	}

	lineData := make([]opts.BarData, len(vadList))
	for i := 0; i < len(vadList); i++ {
		value := -1
		if vadList[i].Active == true {
			value = 1
		}
		lineData[i] = opts.BarData{Value: value}
	}

	// Put data into instance
	line.SetXAxis(xAxis).
		AddSeries("VAD", lineData)

	// Where the magic happens
	f, err := os.Create(filepath.Join(pkg.DefDebugFolder, title+".html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return line.Render(f)
}
