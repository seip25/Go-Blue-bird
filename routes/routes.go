package routes

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/config"
	"github.com/seip25/Go-Blue-bird/public"
	"github.com/seip25/Go-Blue-bird/routes/api"
	"github.com/seip25/Go-Blue-bird/routes/web"
)

func SetupRouter() *gin.Engine {
	var r *gin.Engine
	if gin.Mode() == gin.ReleaseMode {
		r = gin.New()
		r.Use(gin.Recovery())
	} else {
		r = gin.Default()
	}

	r.Use(corsMiddleware())
	r.Use(securityHeadersMiddleware())

	subPublic, err := fs.Sub(public.FS, ".")
	if err == nil {
		publicGroup := r.Group("/public")
		publicGroup.Use(staticCacheMiddleware())
		publicGroup.StaticFS("/", http.FS(subPublic))
	}

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

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		c.Next()
	}
}

func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.IsProd() {
			c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Writer.Header().Set("Pragma", "no-cache")
			c.Writer.Header().Set("Expires", "0")
		}
		c.Next()
	}
}
