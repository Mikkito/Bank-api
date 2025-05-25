package main

import (
	"log"

	"bank-api/cmd/server"
	"bank-api/internal/config"
	"bank-api/migrations"

	_ "github.com/lib/pq"
)

func main() {
	// Загружаем конфиг
	err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Запускаем миграции
	migrations.Run(config.AppConfig)

	// Запускаем сервер, внутри которого уже инициализируется DB и роутинг
	server.StartServer()
}

func migrate(cfg *config.Config) {
	migrations.Run(cfg)
}
