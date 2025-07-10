package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
)

type CefLog struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type Node struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"` // 0-src, 1-dst
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

var (
	originalData GraphData
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Загрузка данных при старте
	loadData()

	// Отдаём index.html
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Эндпоинт отдачи всего графа или фильтрованного по src
	r.GET("/graph", func(c *gin.Context) {
		src := c.Query("src")
		if src == "" {
			// Отдаём весь граф
			c.JSON(http.StatusOK, originalData)
			return
		}

		// Фильтруем граф по src
		graph := filterGraphBySrc(src)
		c.JSON(http.StatusOK, graph)
	})

	// Эндпоинт для списка уникальных src
	r.GET("/src_list", func(c *gin.Context) {
		srcList := uniqueSrcList()
		c.JSON(http.StatusOK, srcList)
	})

	r.Run(":8080")
}

func loadData() {
	path := filepath.Join("dataset", "cef_logs.json")
	file, err := os.ReadFile(path)
	if err != nil {
		panic("Failed to read dataset/cef_logs.json: " + err.Error())
	}

	var logs []CefLog
	if err := json.Unmarshal(file, &logs); err != nil {
		panic("Failed to parse JSON: " + err.Error())
	}

	nodeMap := make(map[string]int) // id -> category
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

	originalData = GraphData{Nodes: nodes, Links: links}
}

func uniqueSrcList() []string {
	srcSet := make(map[string]struct{})
	for _, n := range originalData.Nodes {
		if n.Category == 0 {
			srcSet[n.Id] = struct{}{}
		}
	}
	srcList := make([]string, 0, len(srcSet))
	for src := range srcSet {
		srcList = append(srcList, src)
	}
	sort.Strings(srcList)
	return srcList
}

func filterGraphBySrc(selectedSrc string) GraphData {
	// Находим dst, связанные с выбранным src
	selectedSrcLinks := []Link{}
	selectedDstSet := make(map[string]struct{})

	for _, l := range originalData.Links {
		if l.Source == selectedSrc {
			selectedSrcLinks = append(selectedSrcLinks, l)
			selectedDstSet[l.Target] = struct{}{}
		}
	}

	filteredNodesMap := make(map[string]Node)
	filteredLinks := []Link{}

	// Добавляем выбранный src
	for _, n := range originalData.Nodes {
		if n.Id == selectedSrc {
			filteredNodesMap[n.Id] = n
			break
		}
	}

	// Добавляем dst узлы и связи выбранного src
	for _, l := range selectedSrcLinks {
		filteredLinks = append(filteredLinks, l)
		for _, n := range originalData.Nodes {
			if n.Id == l.Target {
				filteredNodesMap[n.Id] = n
				break
			}
		}
	}

	// Находим связанные src, у которых есть общий dst с выбранным src
	relatedSrcSet := make(map[string]struct{})
	for dst := range selectedDstSet {
		for _, l := range originalData.Links {
			if l.Target == dst && l.Source != selectedSrc {
				relatedSrcSet[l.Source] = struct{}{}
			}
		}
	}

	// Добавляем в граф все связи связанных src (полные), узлы src и dst
	for relatedSrc := range relatedSrcSet {
		// Добавляем src-узел
		for _, n := range originalData.Nodes {
			if n.Id == relatedSrc {
				filteredNodesMap[n.Id] = n
				break
			}
		}
		// Добавляем все связи этого src
		for _, l := range originalData.Links {
			if l.Source == relatedSrc {
				filteredLinks = append(filteredLinks, l)
				// Добавляем dst-узел
				for _, n := range originalData.Nodes {
					if n.Id == l.Target {
						filteredNodesMap[n.Id] = n
						break
					}
				}
			}
		}
	}

	// Уникализируем связи (по ключу source->target)
	uniqueLinksMap := make(map[string]Link)
	for _, l := range filteredLinks {
		key := l.Source + "->" + l.Target
		uniqueLinksMap[key] = l
	}

	nodes := make([]Node, 0, len(filteredNodesMap))
	for _, n := range filteredNodesMap {
		nodes = append(nodes, n)
	}

	return GraphData{
		Nodes: nodes,
		Links: func() []Link {
			res := make([]Link, 0, len(uniqueLinksMap))
			for _, l := range uniqueLinksMap {
				res = append(res, l)
			}
			return res
		}(),
	}
}
