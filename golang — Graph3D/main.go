package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Node struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Size     float64 `json:"symbolSize"`
	Category int     `json:"category"`
	Value    int     `json:"value"`
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
			Value:    1,
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

func main() {
	router := gin.Default()
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		c.File("./assets/index.html")
	})

	router.GET("/graph-data", func(c *gin.Context) {
		data := generateGraph(100, 5)
		c.JSON(http.StatusOK, data)
	})

	router.Run(":8080")
}
