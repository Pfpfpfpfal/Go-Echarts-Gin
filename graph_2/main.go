package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// Структура JSON
type GraphData struct {
	Nodes int      `json:"nodes"`
	Edges [][2]int `json:"edges"`
	Label int      `json:"label"`
}

// Утилита
func floatsToInterfaces(f []float64) []interface{} {
	out := make([]interface{}, len(f))
	for i, v := range f {
		out[i] = v
	}
	return out
}

// Загружаем граф из файла
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

// Генерация случайных координат узлов
func generateNodePositions(n int) map[int][]float64 {
	rand.Seed(time.Now().UnixNano())
	positions := make(map[int][]float64)
	for i := 0; i < n; i++ {
		positions[i] = []float64{
			rand.Float64() * 100,
			rand.Float64() * 100,
			rand.Float64() * 100,
		}
	}
	return positions
}

func create3DGraphChart(graph *GraphData) *charts.Line3D {
	chart := charts.NewLine3D()

	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: "3D-граф",
			Width:     "100%",
			Height:    "800px",
		}),
	)

	positions := generateNodePositions(graph.Nodes)

	// Отображение всех рёбер как отрезков
	var edgesData []opts.Chart3DData
	for _, edge := range graph.Edges {
		from := positions[edge[0]]
		to := positions[edge[1]]

		// Линия между двумя точками — как две точки подряд
		edgesData = append(edgesData,
			opts.Chart3DData{Value: floatsToInterfaces(from)},
			opts.Chart3DData{Value: floatsToInterfaces(to)},
		)
	}

	// Отображение всех узлов как точек
	var nodesData []opts.Chart3DData
	for i := 0; i < graph.Nodes; i++ {
		nodesData = append(nodesData, opts.Chart3DData{
			Value: floatsToInterfaces(positions[i]),
			Name:  fmt.Sprintf("Node %d", i),
		})
	}

	// Добавляем линии (все рёбра)
	chart.AddSeries("Edges", edgesData)

	// Добавляем точки
	chart.AddSeries("Nodes", nodesData,
		func(s *charts.SingleSeries) {
			s.Type = types.ChartScatter3D
			s.SymbolSize = 10
			s.ItemStyle = &opts.ItemStyle{Color: "red"}
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
			Title:    "3D-граф из JSON",
			Subtitle: fmt.Sprintf("Label: %d", graph.Label),
		}),
		charts.WithGrid3DOpts(opts.Grid3D{
			ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(false)},
		}),
	)

	return chart
}

func generateHTMLPage(n int) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>PROTEINS Graph Viewer</title>
	<style>
		body { margin: 0; display: flex; font-family: sans-serif; }
		.sidebar {
			width: 250px;
			background: #f5f5f5;
			overflow-y: scroll;
			height: 100vh;
			border-right: 1px solid #ddd;
			padding: 10px;
		}
		.sidebar a {
			display: block;
			padding: 8px;
			margin: 4px 0;
			text-decoration: none;
			color: #333;
			border-radius: 4px;
		}
		.sidebar a:hover {
			background: #ddd;
		}
		iframe {
			flex-grow: 1;
			height: 100vh;
			border: none;
		}
	</style>
</head>
<body>
	<div class="sidebar">`

	for i := 0; i < n; i++ {
		html += fmt.Sprintf(`<a href="/graph/%d" target="graphFrame">Graph %d</a>`, i, i)
	}

	html += `</div>
	<iframe name="graphFrame" src="/graph/0"></iframe>
</body>
</html>`
	return html
}

func main() {
	graphs, err := loadGraphs("dataset/test.json")
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// Главная страница — список графов + iframe
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, generateHTMLPage(len(graphs)))
	})

	// Отдаём граф по индексу
	r.GET("/graph/:id", func(c *gin.Context) {
		index := c.Param("id")
		var i int
		fmt.Sscanf(index, "%d", &i)
		if i < 0 || i >= len(graphs) {
			c.String(400, "некорректный индекс")
			return
		}
		chart := create3DGraphChart(&graphs[i])
		chart.Render(c.Writer)
	})

	r.Run(":8080")
}
