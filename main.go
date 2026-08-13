package main

import (
	"fmt"
	"log"

	"github.com/seip25/Go-Blue-bird/config"
	"github.com/seip25/Go-Blue-bird/core"
	"github.com/seip25/Go-Blue-bird/routes"
)

func main() {
	cfg := config.Load()

	db, err := core.ConnectDB(cfg.DB)
	if err != nil {
		log.Printf("Warning: Database initialization failed: %v", err)
	} else {
		defer db.Close()
	}

	router := routes.SetupRouter()

	address := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server running on http://localhost%s (Mode: %s)", address, cfg.AppEnv)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
