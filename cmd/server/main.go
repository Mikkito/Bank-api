package server

import (
	"bank-api/internal/api"
	"bank-api/internal/config"
	"bank-api/pkg/utils/logger"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func StartServer() {
	// Logger initialization
	err := logger.Init("pkg/utils/logger/config.yaml")
	if err != nil {
		log.Fatalf("Could not initialize logger: %v", err)
	}
	defer logger.Sync()
	logger := logger.Sugared()

	// Load config
	err = config.Load("configs/config.yaml")
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	// DB initialization
	db := initDB(config.AppConfig)

	// Server initialization
	srv := NewServer()

	// Register routes with DB
	api.RegisterRoutes(srv.Router, db)

	// HTTP-server start
	logger.Info("Starting server on :8080...")
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func initDB(cfg *config.Config) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	return db
}
