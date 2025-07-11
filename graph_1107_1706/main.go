package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type CefLog struct {
	Src   string `json:"src"`
	Dst   string `json:"dst"`
	Dpt   int    `json:"dpt"`
	Start int64  `json:"start"` // UNIX timestamp в миллисекундах
}

type Node struct {
	Id       string `json:"id"`
	Category int    `json:"category"` // 0 - src, 1 - dst
	Dpt      int    `json:"dpt"`      // порт назначения для dst, 0 для src
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

var (
	logs        []CefLog
	allSrcList  []string
	minStart    int64
	maxStart    int64
	dstToSrcMap map[string]map[string]struct{}
)

func main() {
	r := gin.Default()

	r.SetHTMLTemplate(template.Must(template.New("").Funcs(template.FuncMap{
		"marshal": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}).ParseFiles("templates/index.html")))

	err := loadData("dataset/cef_logs.json")
	if err != nil {
		panic(err)
	}

	r.GET("/", func(c *gin.Context) {
		selectedSrc := c.Query("src")
		startFromStr := c.Query("start_from")
		startToStr := c.Query("start_to")
		commonOnly := c.Query("common") == "on"

		const layout = "2006-01-02T15:04"

		var startFrom, startTo int64
		if t, err := time.Parse(layout, startFromStr); err == nil {
			startFrom = t.UnixMilli()
		}
		if t, err := time.Parse(layout, startToStr); err == nil {
			startTo = t.UnixMilli()
		}

		nodes, links := filterGraph(selectedSrc, startFrom, startTo, commonOnly)
		nodesJSON, _ := json.Marshal(nodes)
		linksJSON, _ := json.Marshal(links)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"SrcList":     allSrcList,
			"SelectedSrc": selectedSrc,
			"StartFrom":   startFromStr,
			"StartTo":     startToStr,
			"MinStart":    time.UnixMilli(minStart).Format(layout),
			"MaxStart":    time.UnixMilli(maxStart).Format(layout),
			"CommonOnly":  commonOnly,
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

	srcSet := make(map[string]struct{})
	dstToSrcMap = make(map[string]map[string]struct{})

	for i, log := range logs {
		srcSet[log.Src] = struct{}{}

		if dstToSrcMap[log.Dst] == nil {
			dstToSrcMap[log.Dst] = make(map[string]struct{})
		}
		dstToSrcMap[log.Dst][log.Src] = struct{}{}

		if i == 0 || log.Start < minStart {
			minStart = log.Start
		}
		if i == 0 || log.Start > maxStart {
			maxStart = log.Start
		}
	}

	for src := range srcSet {
		allSrcList = append(allSrcList, src)
	}
	sort.Strings(allSrcList)
	return nil
}

func filterGraph(selectedSrc string, startFrom, startTo int64, commonOnly bool) ([]Node, []Link) {
	srcFilter := make(map[string]struct{})

	for _, log := range logs {
		if startFrom > 0 && log.Start < startFrom {
			continue
		}
		if startTo > 0 && log.Start > startTo {
			continue
		}
		if selectedSrc == "" || log.Src == selectedSrc {
			srcFilter[log.Src] = struct{}{}
		}
	}

	if commonOnly {
		shared := make(map[string]struct{})
		for _, srcs := range dstToSrcMap {
			if len(srcs) > 1 {
				for s := range srcs {
					shared[s] = struct{}{}
				}
			}
		}
		for s := range srcFilter {
			if _, ok := shared[s]; !ok {
				delete(srcFilter, s)
			}
		}
	}

	nodeMap := make(map[string]Node)
	linkMap := make(map[string]Link)

	for _, log := range logs {
		if startFrom > 0 && log.Start < startFrom {
			continue
		}
		if startTo > 0 && log.Start > startTo {
			continue
		}
		if _, ok := srcFilter[log.Src]; !ok {
			continue
		}
		if _, exists := nodeMap[log.Src]; !exists {
			nodeMap[log.Src] = Node{Id: log.Src, Category: 0, Dpt: 0}
		}
		if _, exists := nodeMap[log.Dst]; !exists {
			nodeMap[log.Dst] = Node{Id: log.Dst, Category: 1, Dpt: log.Dpt}
		}
		linkMap[log.Src+"->"+log.Dst] = Link{Source: log.Src, Target: log.Dst}
	}

	nodes := make([]Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	links := make([]Link, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}

	return nodes, links
}
