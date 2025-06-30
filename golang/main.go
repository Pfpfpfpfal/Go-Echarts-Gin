package main

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Одна линия — набор точек [x,y,z]
type Line [][]float64

func generateLine(tMax, phase, scale float64) Line {
	const step = 0.01
	var line Line
	for t := 0.0; t < tMax; t += step {
		x := scale * (1 + 0.25*math.Cos(75*t+phase)) * math.Cos(t)
		y := scale * (1 + 0.25*math.Cos(75*t+phase)) * math.Sin(t)
		z := scale * (t + 2.0*math.Sin(75*t+phase))
		line = append(line, []float64{x, y, z})
	}
	return line
}

func main() {
	rand.Seed(time.Now().UnixNano())

	router := gin.Default()
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		c.File("./assets/index.html")
	})

	router.GET("/lines", func(c *gin.Context) {
		nLines := 3
		lines := make([]Line, 0, nLines)
		for i := 0; i < nLines; i++ {
			phase := float64(i) * 1.5
			scale := 1.0 + 0.5*float64(i)
			lines = append(lines, generateLine(25, phase, scale))
		}
		c.JSON(http.StatusOK, lines)
	})

	router.Run(":8080")
}
