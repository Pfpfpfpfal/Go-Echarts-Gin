package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Node struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Size     float64 `json:"symbolSize"`
	Category int     `json:"category"`
	Value    float32 `json:"value"`
	ID       int     `json:"id"`
	Name     string  `json:"name"`
}

type Edge struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Value  int `json:"value"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func generateGraph(numNodes int, numCategories int) GraphData {
	rand.Seed(time.Now().UnixNano())

	nodes := make([]Node, numNodes)
	edges := make([]Edge, 0)

	for i := 0; i < numNodes; i++ {
		nodes[i] = Node{
			X:        rand.Float64() * 1000,
			Y:        rand.Float64() * 1000,
			Size:     rand.Float64()*10 + 5,
			Category: rand.Intn(numCategories),
			Value:    float32(1),
			ID:       i,
			Name:     "Node " + string(rune(i+65)),
		}
	}

	for i := 0; i < numNodes; i++ {
		for j := 0; j < rand.Intn(3)+1; j++ {
			target := rand.Intn(numNodes)
			if target != i {
				edges = append(edges, Edge{Source: i, Target: target, Value: 2})
			}
		}
	}

	return GraphData{Nodes: nodes, Edges: edges}
}

func graphGL() *charts.Graph {
	data := generateGraph(100, 5)

	// Convert nodes to go-echarts format
	graphNodes := make([]opts.GraphNode, len(data.Nodes))
	for i, n := range data.Nodes {
		graphNodes[i] = opts.GraphNode{
			Name:       n.Name,
			X:          float32(n.X),
			Y:          float32(n.Y),
			SymbolSize: n.Size,
			Category:   n.Category,
			Value:      float32(n.Value),
		}
	}

	// Convert edges to go-echarts format
	graphLinks := make([]opts.GraphLink, len(data.Edges))
	for i, e := range data.Edges {
		sourceName := data.Nodes[e.Source].Name
		targetName := data.Nodes[e.Target].Name
		graphLinks[i] = opts.GraphLink{
			Source: sourceName,
			Target: targetName,
			Value:  float32(e.Value),
		}
	}

	// Create categories
	categories := make([]*opts.GraphCategory, 0)
	categoriesMap := make(map[int]bool)
	for _, n := range data.Nodes {
		if !categoriesMap[n.Category] {
			categories = append(categories, &opts.GraphCategory{
				Name: "Category " + string(rune(n.Category+65)),
			})
			categoriesMap[n.Category] = true
		}
	}

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Graph with ForceAtlas2"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: boolptr(true)}),
		charts.WithLegendOpts(opts.Legend{Show: boolptr(true)}),
	)

	graph.AddSeries("graph", graphNodes, graphLinks,
		charts.WithGraphChartOpts(
			opts.GraphChart{
				Force: &opts.GraphForce{
					Repulsion:  8000,
					Gravity:    0.1,
					EdgeLength: 100,
				},
				Categories:         categories,
				Roam:               boolptr(true),
				FocusNodeAdjacency: boolptr(true),
			},
		),
		charts.WithEmphasisOpts(
			opts.Emphasis{
				Label: &opts.Label{
					Show:     boolptr(true),
					Color:    "black",
					Position: "right",
				},
			},
		),
		charts.WithLineStyleOpts(
			opts.LineStyle{
				Color: "rgba(255,255,255,0.2)",
			},
		),
		charts.WithItemStyleOpts(
			opts.ItemStyle{
				Opacity: opts.Float(1.0),
			},
		),
	)

	return graph
}

func boolptr(b bool) *bool {
	return &b
}

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		page := components.NewPage()
		page.AddCharts(graphGL())

		// Render the page directly to the response writer
		err := page.Render(c.Writer)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	router.Run(":8080")
}
