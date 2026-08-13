package routes

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/public"
	"github.com/seip25/Go-Blue-bird/routes/api"
	"github.com/seip25/Go-Blue-bird/routes/web"
	"github.com/seip25/Go-Blue-bird/templates"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	subPublic, err := fs.Sub(public.FS, ".")
	if err == nil {
		r.StaticFS("/public", http.FS(subPublic))
	}

	tmpl := template.Must(template.New("").ParseFS(templates.FS, "layouts/*.html", "pages/*.html"))
	r.SetHTMLTemplate(tmpl)

	webGroup := r.Group("/")
	web.RegisterRoutes(webGroup)

	apiGroup := r.Group("/api")
	api.RegisterRoutes(apiGroup)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
