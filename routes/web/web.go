package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/", HomeHandler)
	rg.GET("/about", AboutHandler)
}

func HomeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title": "Go Blue Bird - Home",
	})
}

func AboutHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", gin.H{
		"Title": "Go Blue Bird - About",
	})
}
