package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"StarstreamAstra/internal/config"
	"StarstreamAstra/internal/db"
	"Zjmf-kvm/internal/router"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var zapLogger *zap.Logger
	if cfg != nil && cfg.Logger.UseZap {
		zapLogger, err = config.NewZapLogger(&cfg.Logger)
		if err != nil {
			log.Fatalf("Failed to create zap logger: %v", err)
		}
		defer func() { _ = zapLogger.Sync() }()
	}

	if cfg != nil && cfg.Database.RunMigrations {
		if cfg.Database.URL == "" {
			log.Fatalf("database.url is empty but run_migrations=true; please set database.url in config")
		}
		if err := db.RunMigrations(cfg.Database.URL, cfg.Database.MigrationsPath); err != nil {
			if zapLogger != nil {
				zapLogger.Sugar().Fatalf("Failed to run migrations: %v", err)
			} else {
				log.Fatalf("Failed to run migrations: %v", err)
			}
		}
	}

	useZap := false
	loggerChoice := "info"
	if cfg != nil {
		useZap = cfg.Logger.UseZap
		loggerChoice = cfg.Database.LoggerLevel
	}
	dbConn, err := db.InitDB(cfg.Database.DSN, useZap, zapLogger, loggerChoice)
	if err != nil {
		if zapLogger != nil {
			zapLogger.Sugar().Fatalf("Failed to init db: %v", err)
		} else {
			log.Fatalf("Failed to init db: %v", err)
		}
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			if zapLogger != nil {
				zapLogger.Sugar().Errorf("Error closing db: %v", err)
			} else {
				log.Printf("Error closing db: %v", err)
			}
		}
	}()

	if cfg != nil && cfg.Server.Port == "" {
		cfg.Server.Port = "1270"
	}
	r := gin.New()
	r.Use(gin.Recovery())
	if zapLogger != nil {
		r.Use(func(c *gin.Context) { c.Next() })
	} else {
		r.Use(gin.Logger())
	}

	router.RegisterRoutes(r, dbConn, cfg)

	port := cfg.Server.Port
	if envPort := os.Getenv("HTTP_PORT"); envPort != "" {
		port = envPort
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	go func() {
		if zapLogger != nil {
			zapLogger.Sugar().Infof("Starting server on %s", srv.Addr)
		} else {
			log.Printf("Starting server on %s", srv.Addr)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if zapLogger != nil {
				zapLogger.Sugar().Fatalf("Listen: %v", err)
			} else {
				log.Fatalf("Listen: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if zapLogger != nil {
		zapLogger.Sugar().Info("Shutting down server...")
	} else {
		log.Println("Shutting down server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		if zapLogger != nil {
			zapLogger.Sugar().Fatalf("Server forced to shutdown: %v", err)
		} else {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}

	if zapLogger != nil {
		zapLogger.Sugar().Info("Server exiting")
	} else {
		log.Println("Server exiting")
	}
}
