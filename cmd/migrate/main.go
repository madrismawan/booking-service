package main

import (
	"errors"
	"flag"
	"log"
	"strings"

	"booking-service/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up")
	path := flag.String("path", "migration", "migration files path")
	force := flag.Bool("force", false, "allow destructive migration commands outside development")
	flag.Parse()

	cfg := config.Load()
	switch *direction {
	case "up":
	case "fresh":
		if !isFreshAllowed(cfg.AppEnv, *force) {
			log.Fatalf("fresh migration is destructive and only allowed when APP_ENV=development or -force=true")
		}
	default:
		log.Fatalf("unsupported migration direction %q, supported directions: up, fresh", *direction)
	}

	m, err := migrate.New("file://"+*path, cfg.DB.URL())
	if err != nil {
		log.Fatalf("create migration: %v", err)
	}
	defer m.Close()

	switch *direction {
	case "up":
		if err := migrateUp(m); err != nil {
			log.Fatalf("run migration: %v", err)
		}
		log.Println("migration up completed")
	case "fresh":
		if err := m.Drop(); err != nil {
			log.Fatalf("drop database objects: %v", err)
		}
		if err := migrateUp(m); err != nil {
			log.Fatalf("run fresh migration: %v", err)
		}
		log.Println("migration fresh completed")
	default:
		log.Fatalf("unsupported migration direction %q, supported directions: up, fresh", *direction)
	}
}

func migrateUp(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func isFreshAllowed(appEnv string, force bool) bool {
	return force || strings.EqualFold(strings.TrimSpace(appEnv), "development")
}
