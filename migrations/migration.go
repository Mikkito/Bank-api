package migrations

import (
	"bank-api/internal/config"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(cfg *config.Config) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.DBName, cfg.Database.SSLMode)

	m, err := migrate.New(cfg.Database.MigrationsPath, dsn)
	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Fatalf("Migration up error: %v", err)
	}
}
