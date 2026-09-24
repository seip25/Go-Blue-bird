package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/config"
	"github.com/seip25/Go-Blue-bird/core"
	"github.com/seip25/Go-Blue-bird/routes"
)

func main() {
	cfg := config.Load()

	if config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := core.ConnectDB(cfg.DB)
	if err != nil {
		log.Printf("Warning: Database initialization failed: %v", err)
	} else {
		defer db.Close()
	}

	router := routes.SetupRouter()

	address := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	go func() {
		log.Printf("Server running on http://localhost%s (Mode: %s)", address, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
