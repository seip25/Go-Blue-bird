package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/config"
	"github.com/seip25/Go-Blue-bird/core"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", HealthHandler)
	rg.HEAD("/health", HealthHandler)
	rg.GET("/info", InfoHandler)
	rg.HEAD("/info", InfoHandler)
	rg.POST("/users", CreateUserSampleHandler)
}

func HealthHandler(c *gin.Context) {
	dbStatus := "connected"
	if err := core.PingDB(); err != nil {
		dbStatus = "disconnected"
	}

	core.RespondSuccess(c, "Service is healthy", gin.H{
		"status":    "up",
		"database":  dbStatus,
		"timestamp": time.Now().UTC(),
	})
}

func InfoHandler(c *gin.Context) {
	cfg := config.GetGlobal()
	core.RespondSuccess(c, "Application information", gin.H{
		"app_name":  cfg.AppName,
		"app_env":   cfg.AppEnv,
		"port":      cfg.Port,
		"db_driver": cfg.DB.Driver,
	})
}

func CreateUserSampleHandler(c *gin.Context) {
	var req CreateUserRequest
	if !core.BindJSONAndValidate(c, &req) {
		return
	}

	hashedPassword, err := core.HashPassword(req.Password)
	if err != nil {
		core.RespondError(c, http.StatusInternalServerError, "Failed to process security payload", err.Error())
		return
	}

	core.RespondCreated(c, "User validated and created successfully", gin.H{
		"username": req.Username,
		"email":    req.Email,
		"hash":     hashedPassword,
	})
}
