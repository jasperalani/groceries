package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)

	var list = []string{
		"noodle",
		"poodle",
		"doodle",
	}

	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.StaticFS("/static", http.Dir("./static"))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Title of list",
			"list":  list,
		})
	})

	router.Run(":8080")
}
