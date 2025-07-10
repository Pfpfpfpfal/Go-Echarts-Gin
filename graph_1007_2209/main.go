package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
)

type CefLog struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type Node struct {
	Id       string `json:"id"`
	Category int    `json:"category"` // 0 - src, 1 - dst
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

var (
	logs          []CefLog
	originalNodes []Node
	originalLinks []Link
	allSrcList    []string
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	err := loadData("dataset/cef_logs.json")
	if err != nil {
		panic(err)
	}

	r.GET("/", func(c *gin.Context) {
		selectedSrc := c.Query("src")

		nodes, links := filterGraphBySrc(selectedSrc)

		nodesJSON, _ := json.Marshal(nodes)
		linksJSON, _ := json.Marshal(links)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"SrcList":     allSrcList,
			"SelectedSrc": selectedSrc,
			"NodesJSON":   template.JS(nodesJSON),
			"LinksJSON":   template.JS(linksJSON),
		})
	})

	r.Run(":8080")
}

func loadData(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &logs)
	if err != nil {
		return err
	}

	nodeMap := make(map[string]int)
	nodes := []Node{}
	links := []Link{}

	for _, log := range logs {
		if _, ok := nodeMap[log.Src]; !ok {
			nodeMap[log.Src] = 0
			nodes = append(nodes, Node{Id: log.Src, Category: 0})
		}
		if _, ok := nodeMap[log.Dst]; !ok {
			nodeMap[log.Dst] = 1
			nodes = append(nodes, Node{Id: log.Dst, Category: 1})
		}
		links = append(links, Link{Source: log.Src, Target: log.Dst})
	}

	originalNodes = nodes
	originalLinks = links

	srcSet := make(map[string]struct{})
	for _, n := range nodes {
		if n.Category == 0 {
			srcSet[n.Id] = struct{}{}
		}
	}
	for src := range srcSet {
		allSrcList = append(allSrcList, src)
	}
	sort.Strings(allSrcList)

	return nil
}

func filterGraphBySrc(selectedSrc string) ([]Node, []Link) {
	if selectedSrc == "" {
		return originalNodes, originalLinks
	}

	dstSet := make(map[string]struct{})
	for _, l := range originalLinks {
		if l.Source == selectedSrc {
			dstSet[l.Target] = struct{}{}
		}
	}

	filteredNodeMap := make(map[string]Node)
	filteredLinks := []Link{}

	// Добавим выбранный src
	for _, n := range originalNodes {
		if n.Id == selectedSrc {
			filteredNodeMap[n.Id] = n
			break
		}
	}

	// Добавим dst и связи выбранного src
	for _, l := range originalLinks {
		if l.Source == selectedSrc {
			filteredLinks = append(filteredLinks, l)
			for _, n := range originalNodes {
				if n.Id == l.Target {
					filteredNodeMap[n.Id] = n
					break
				}
			}
		}
	}

	// Найдем другие src с общими dst
	relatedSrcSet := make(map[string]struct{})
	for dst := range dstSet {
		for _, l := range originalLinks {
			if l.Target == dst && l.Source != selectedSrc {
				relatedSrcSet[l.Source] = struct{}{}
			}
		}
	}

	// Добавим эти src и их связи
	for relatedSrc := range relatedSrcSet {
		for _, n := range originalNodes {
			if n.Id == relatedSrc {
				filteredNodeMap[n.Id] = n
				break
			}
		}
		for _, l := range originalLinks {
			if l.Source == relatedSrc {
				filteredLinks = append(filteredLinks, l)
				for _, n := range originalNodes {
					if n.Id == l.Target {
						filteredNodeMap[n.Id] = n
						break
					}
				}
			}
		}
	}

	// Уникализируем связи
	linkMap := make(map[string]Link)
	for _, l := range filteredLinks {
		key := l.Source + "->" + l.Target
		linkMap[key] = l
	}

	nodes := make([]Node, 0, len(filteredNodeMap))
	for _, n := range filteredNodeMap {
		nodes = append(nodes, n)
	}

	uniqueLinks := make([]Link, 0, len(linkMap))
	for _, l := range linkMap {
		uniqueLinks = append(uniqueLinks, l)
	}

	return nodes, uniqueLinks
}
