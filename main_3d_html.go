package main

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		const page = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>3D график</title>
    <!-- ECharts и ECharts-GL -->
    <script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/echarts-gl@2/dist/echarts-gl.min.js"></script>
</head>
<body>
    <div id="chart" style="width: 100%; height: 600px;"></div>
    <script>
        var chart = echarts.init(document.getElementById('chart'));

        var lineData = [
            [1, 3, 2],
            [2, 5, 3],
            [3, 4, 5],
            [4, 6, 4],
            [5, 7, 6]
        ];

        var scatterData = [
            [1.5, 3.5, 2.5],
            [2.5, 4.5, 3.2],
            [3.5, 5,   5.5],
            [4.5, 6.5, 4.8]
        ];

        chart.setOption({
            tooltip: {},
            xAxis3D: { type: 'value' },
            yAxis3D: { type: 'value' },
            zAxis3D: { type: 'value' },
            grid3D: {
                viewControl: {
                    autoRotate: true
                }
            },
            series: [
                {
                    name: "Линия",
                    type: 'line3D',
                    data: lineData,
                    lineStyle: {
                        width: 4
                    }
                },
                {
                    name: "Точки",
                    type: 'scatter3D',
                    data: scatterData,
                    symbolSize: 12
                }
            ]
        });
    </script>
</body>
</html>
`
		tmpl := template.Must(template.New("page").Parse(page))
		c.Status(http.StatusOK)
		tmpl.Execute(c.Writer, nil)
	})

	r.Run(":8080")
}
