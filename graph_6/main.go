package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// Структура JSON
type GraphData struct {
	Nodes     int         `json:"nodes"`
	Edges     [][2]int    `json:"edges"`
	Label     int         `json:"label"`
	Positions [][]float64 `json:"positions"`
}

// Утилита
func floatsToInterfaces(f []float64) []interface{} {
	out := make([]interface{}, len(f))
	for i, v := range f {
		out[i] = v
	}
	return out
}

// Загрузка графов
func loadGraphs(path string) ([]GraphData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graphs []GraphData
	if err := json.Unmarshal(data, &graphs); err != nil {
		return nil, err
	}
	if len(graphs) == 0 {
		return nil, fmt.Errorf("массив графов пустой")
	}
	return graphs, nil
}

// Генерация графа в формате 3D
func create3DGraphChart(graph *GraphData, index int) *charts.Line3D {
	chart := charts.NewLine3D()

	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: "3D-граф",
			Width:     "100%",
			Height:    "800px",
		}),
	)

	positions := make(map[int][]float64)
	for i := 0; i < graph.Nodes; i++ {
		positions[i] = graph.Positions[i]
	}

	for _, edge := range graph.Edges {
		from := positions[edge[0]]
		to := positions[edge[1]]

		chart.AddSeries(fmt.Sprintf("Edge %d-%d", edge[0], edge[1]),
			[]opts.Chart3DData{
				{Value: floatsToInterfaces(from)},
				{Value: floatsToInterfaces(to)},
			},
			func(s *charts.SingleSeries) {
				s.Type = types.ChartLine3D
				s.ItemStyle = &opts.ItemStyle{Color: "gray"}
			},
		)
	}

	var nodesData []opts.Chart3DData
	for i := 0; i < graph.Nodes; i++ {
		nodesData = append(nodesData, opts.Chart3DData{
			Value: floatsToInterfaces(positions[i]),
			Name:  fmt.Sprintf("Node %d", i),
		})
	}

	chart.AddSeries("Nodes", nodesData,
		func(s *charts.SingleSeries) {
			s.Type = types.ChartScatter3D
			s.SymbolSize = 10
			s.ItemStyle = &opts.ItemStyle{Color: "blue"}
			s.Label = &opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{b}",
				Color:     "black",
				Position:  "top",
			}
		},
	)

	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("Graph %d", index),
			Subtitle: fmt.Sprintf("Label: %d", graph.Label),
		}),
		charts.WithGrid3DOpts(opts.Grid3D{
			ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(false)},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false),
		}),
	)

	return chart
}

func main() {
	graphs, err := loadGraphs("dataset/test.json")
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// Главная страница — поле ввода и iframe
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Graph Viewer</title>
	<script>
		function showGraph() {
			const id = document.getElementById("graphInput").value;
			document.getElementById("graphFrame").src = "/graph/" + id;
		}
	</script>
</head>
<body style="margin:0; font-family:sans-serif">
	<div style="padding:10px; background:#eee; border-bottom:1px solid #ccc">
		<label>Граф №</label>
		<input id="graphInput" type="number" min="0" max="`+fmt.Sprintf("%d", len(graphs)-1)+`" value="0">
		<button onclick="showGraph()">Показать</button>
	</div>
	<iframe id="graphFrame" src="/graph/0" style="width:100%; height:90vh; border:none;"></iframe>
</body>
</html>`)
	})

	r.GET("/graph/:id", func(c *gin.Context) {
		index := c.Param("id")
		var i int
		fmt.Sscanf(index, "%d", &i)
		if i < 0 || i >= len(graphs) {
			c.String(400, "некорректный индекс")
			return
		}
		chart := create3DGraphChart(&graphs[i], i)
		chart.Render(c.Writer)
	})

	r.Run(":8080")
}
