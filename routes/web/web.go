package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/core"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/", HomeHandler)
	rg.HEAD("/", HomeHandler)
	rg.GET("/about", AboutHandler)
	rg.HEAD("/about", AboutHandler)
}

func HomeHandler(c *gin.Context) {
	core.Render(c, http.StatusOK, "index", gin.H{
		"Title": "Go Blue Bird - Home",
	})
}

func AboutHandler(c *gin.Context) {
	core.Render(c, http.StatusOK, "about", gin.H{
		"Title": "Go Blue Bird - About",
	})
}
