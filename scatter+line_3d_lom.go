package main

import (
	"math"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func floatsToInterfaces(f []float64) []interface{} {
	out := make([]interface{}, len(f))
	for i, v := range f {
		out[i] = v
	}
	return out
}

func generateLineData() []opts.Chart3DData {
	data := make([]opts.Chart3DData, 0)
	for x := 0.0; x <= 10; x += 1 {
		y := math.Sin(x) * 5
		z := math.Cos(x) * 5
		data = append(data, opts.Chart3DData{Value: floatsToInterfaces([]float64{x, y, z})})
	}
	return data
}

func generateScatterData() []opts.Chart3DData {
	return []opts.Chart3DData{
		{Value: floatsToInterfaces([]float64{1.5, 3.5, 4.2})},
		{Value: floatsToInterfaces([]float64{2.5, -2.5, -4.5})},
		{Value: floatsToInterfaces([]float64{4, 1.8, -1})},
		{Value: floatsToInterfaces([]float64{7.2, 4.3, 3.1})},
		{Value: floatsToInterfaces([]float64{8.5, -1.2, -3.5})},
	}
}

// Общий набор координат для линии и точек
func generateShared3DPoints() []opts.Chart3DData {
	points := [][]float64{
		{0, 0, 0},
		{1, 2, 1},
		{2, 1, 3},
		{3, 3, 2},
		{4, 0, 4},
	}

	result := make([]opts.Chart3DData, 0, len(points))
	for _, p := range points {
		result = append(result, opts.Chart3DData{
			Value: floatsToInterfaces(p),
		})
	}
	return result
}

// Тут была замена (Кривые точки кривые линии используют genscat and genline)
func createCombined3DChart() *charts.Line3D {
	sharedPoints := generateShared3DPoints()

	line := charts.NewLine3D()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "3D: Линия и точки в одинаковых координатах"}),
		charts.WithGrid3DOpts(opts.Grid3D{
			ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(true)},
		}),
	)

	// Серия линии
	line.AddSeries("Линия", sharedPoints)

	// Серия точек — точно такие же координаты
	line.AddSeries("Точки", sharedPoints,
		func(s *charts.SingleSeries) {
			s.Type = types.ChartScatter3D
			s.SymbolSize = 15
			s.ItemStyle = &opts.ItemStyle{Color: "red"}
		},
	)

	return line
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		chart := createCombined3DChart()
		c.Header("Content-Type", "text/html")
		chart.Render(c.Writer)
	})
	r.Run(":8080")
}
