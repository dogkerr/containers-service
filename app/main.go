package main

import (
	"containerServiceLintang/config"
	"containerServiceLintang/internal/repository/postgres"
	"containerServiceLintang/pkg/gorm"
	"containerServiceLintang/pkg/httpserver"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)



func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	//  databae
	gorm, err := gorm.NewGorm(cfg.Postgres.Username, cfg.Postgres.Password)
	if err != nil {
		log.Fatalf("Database Connection error: %s", err)
	}

	// HTTP Server
	handler := gin.New()
	httpServer := httpserver.New(handler, httpserver.Port(":9090"))


	// Prepare Repository
	containerRepo := postgres.NewContainerRepo(gorm.Pool)

	// Build service layer
	


	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Fatal("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Fatal(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	

}
