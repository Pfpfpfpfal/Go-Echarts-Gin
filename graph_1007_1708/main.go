package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type CefLog struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type Node struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

func main() {
	r := gin.Default()

	// Загрузка шаблонов из папки templates
	r.LoadHTMLGlob("templates/*")

	// Рендерим index.html
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Отдаем граф из JSON файла
	r.GET("/graph", func(c *gin.Context) {
		// Полный путь к файлу
		path := filepath.Join("dataset", "cef_logs.json")

		file, err := os.ReadFile(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cef_logs.json"})
			return
		}

		var logs []CefLog
		if err := json.Unmarshal(file, &logs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
			return
		}

		nodeMap := make(map[string]int) // категории: 0-src, 1-dst
		nodes := []Node{}
		links := []Link{}

		for _, log := range logs {
			if _, ok := nodeMap[log.Src]; !ok {
				nodeMap[log.Src] = 0
				nodes = append(nodes, Node{Id: log.Src, Name: log.Src, Category: 0})
			}
			if _, ok := nodeMap[log.Dst]; !ok {
				nodeMap[log.Dst] = 1
				nodes = append(nodes, Node{Id: log.Dst, Name: log.Dst, Category: 1})
			}
			links = append(links, Link{Source: log.Src, Target: log.Dst, Value: 1})
		}

		graph := GraphData{Nodes: nodes, Links: links}
		c.JSON(http.StatusOK, graph)
	})

	r.Run(":8080")
}
