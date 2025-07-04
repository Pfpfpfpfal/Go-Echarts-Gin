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
func loadGraph(path string) (*GraphData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graphs []GraphData // Массив, а не один объект!
	if err := json.Unmarshal(data, &graphs); err != nil {
		return nil, err
	}
	if len(graphs) == 0 {
		return nil, fmt.Errorf("массив графов пустой")
	}
	return &graphs[0], nil // Берём первый граф
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

func main() {
	graph, err := loadGraph("dataset/test.json")
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		chart := create3DGraphChart(graph)
		chart.Render(c.Writer)
	})
	r.Run(":8080")
}
