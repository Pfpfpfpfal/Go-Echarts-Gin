package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type GraphData struct {
	Nodes     int         `json:"nodes"`
	Edges     [][2]int    `json:"edges"`
	Label     int         `json:"label"`
	Positions [][]float64 `json:"positions"`
}

func floatsToInterfaces(f []float64) []interface{} {
	out := make([]interface{}, len(f))
	for i, v := range f {
		out[i] = v
	}
	return out
}

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

func create3DGraphChart(graph *GraphData, index int, showLabels bool, focusNode int, grouped bool, filterGroup int) *charts.Line3D {
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

	visibleNodes := map[int]bool{}
	if focusNode >= 0 {
		visibleNodes[focusNode] = true
		for _, edge := range graph.Edges {
			if edge[0] == focusNode {
				visibleNodes[edge[1]] = true
			}
			if edge[1] == focusNode {
				visibleNodes[edge[0]] = true
			}
		}
	}

	colors := map[int]string{
		0: "red",
		1: "green",
		2: "blue",
	}

	// Функция для определения группы узла
	groupOf := func(i int) int {
		return i % 3
	}

	// Добавляем рёбра — фильтруем по фокусу и фильтру по группе
	for _, edge := range graph.Edges {
		if focusNode >= 0 && edge[0] != focusNode && edge[1] != focusNode {
			continue
		}
		if grouped && filterGroup >= 0 {
			// Проверяем, что хотя бы один из концов ребра в выбранной группе
			g0 := groupOf(edge[0])
			g1 := groupOf(edge[1])
			if g0 != filterGroup && g1 != filterGroup {
				continue
			}
		}
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

	if grouped {
		// Группируем узлы по цветам
		groupNodes := map[int][]opts.Chart3DData{
			0: {},
			1: {},
			2: {},
		}
		for i := 0; i < graph.Nodes; i++ {
			if focusNode >= 0 && !visibleNodes[i] {
				continue
			}
			g := groupOf(i)
			// Если выбран фильтр группы, пропускаем остальные группы
			if filterGroup >= 0 && g != filterGroup {
				continue
			}
			groupNodes[g] = append(groupNodes[g], opts.Chart3DData{
				Value: floatsToInterfaces(positions[i]),
				Name:  fmt.Sprintf("Node %d", i),
			})
		}

		for group, data := range groupNodes {
			// Добавляем серии с цветом группы
			chart.AddSeries(fmt.Sprintf("Group %d", group), data,
				func(s *charts.SingleSeries) {
					s.Type = types.ChartScatter3D
					s.SymbolSize = 10
					s.ItemStyle = &opts.ItemStyle{Color: colors[group]}
					if showLabels {
						s.Label = &opts.Label{
							Show:      opts.Bool(true),
							Formatter: "{b}",
							Color:     "black",
							Position:  "top",
						}
					} else {
						s.Label = &opts.Label{Show: opts.Bool(false)}
					}
				},
			)
		}
	} else {
		// Без группировки - все узлы синие
		var nodesData []opts.Chart3DData
		for i := 0; i < graph.Nodes; i++ {
			if focusNode >= 0 && !visibleNodes[i] {
				continue
			}
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
				if showLabels {
					s.Label = &opts.Label{
						Show:      opts.Bool(true),
						Formatter: "{b}",
						Color:     "black",
						Position:  "top",
					}
				} else {
					s.Label = &opts.Label{Show: opts.Bool(false)}
				}
			},
		)
	}

	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("Graph %d", index),
			Subtitle: fmt.Sprintf("Label: %d", graph.Label),
		}),
		charts.WithGrid3DOpts(opts.Grid3D{
			ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(false)},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false), // легенду отключаем
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

	tmpl := template.Must(template.ParseFiles(filepath.Join("templates", "index.html")))

	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(c.Writer, gin.H{
			"MaxGraph": len(graphs) - 1,
		})
	})

	r.GET("/graph/:id", func(c *gin.Context) {
		indexStr := c.Param("id")
		i, err := strconv.Atoi(indexStr)
		if err != nil || i < 0 || i >= len(graphs) {
			c.String(400, "некорректный индекс")
			return
		}

		showLabels := true
		if val := c.Query("labels"); val == "0" {
			showLabels = false
		}

		focusNode := -1
		if val := c.Query("focus"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				focusNode = n
			}
		}

		grouped := false
		if val := c.Query("grouped"); val == "1" {
			grouped = true
		}

		filterGroup := -1
		if val := c.Query("filterGroup"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				filterGroup = n
			}
		}

		chart := create3DGraphChart(&graphs[i], i, showLabels, focusNode, grouped, filterGroup)
		chart.Render(c.Writer)
	})

	r.Run(":8080")
}
