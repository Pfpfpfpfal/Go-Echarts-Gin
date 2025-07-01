package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func createMixedChart() *charts.Line {
	// 1) Создаём Line
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Линия + Точки (Overlap)",
			Subtitle: "Объединённый график через Overlap",
		}),
		// делаем числовые оси
		charts.WithXAxisOpts(opts.XAxis{Type: "value"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value"}),
	)

	// Данные линии
	lineData := []opts.LineData{
		{Value: []float64{1, 3}},
		{Value: []float64{2, 5}},
		{Value: []float64{3, 4}},
		{Value: []float64{4, 6}},
		{Value: []float64{5, 7}},
	}
	line.AddSeries("Ломаная линия", lineData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(false)}),
		)

	// 2) Создаём Scatter
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		// Тот же заголовок не важен, главное оси
		charts.WithXAxisOpts(opts.XAxis{Type: "value"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value"}),
	)
	scatterData := []opts.ScatterData{
		{Value: []float64{1.5, 3.5}},
		{Value: []float64{2.5, 4.5}},
		{Value: []float64{3.5, 5}},
		{Value: []float64{4.5, 6.5}},
	}
	scatter.AddSeries("Точки", scatterData).
		SetSeriesOptions(
			charts.WithScatterChartOpts(opts.ScatterChart{SymbolSize: 20}),
		)

	// 3) Накладываем точки на линию
	line.Overlap(scatter)

	return line
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		chart := createMixedChart()
		c.Header("Content-Type", "text/html")
		chart.Render(c.Writer)
	})
	r.Run(":8080")
}
