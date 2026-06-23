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

	switch *direction {
	case "up":
		m, err := newMigrate(*path, cfg.DB.URL())
		if err != nil {
			log.Fatalf("create migration: %v", err)
		}
		defer m.Close()

		if err := migrateUp(m); err != nil {
			log.Fatalf("run migration: %v", err)
		}
		log.Println("migration up completed")
	case "fresh":
		m, err := newMigrate(*path, cfg.DB.URL())
		if err != nil {
			log.Fatalf("create migration: %v", err)
		}

		if err := m.Drop(); err != nil {
			if !isMissingSchemaMigrations(err) {
				log.Fatalf("drop database objects: %v", err)
			}
			log.Printf("skip drop step: %v", err)
		}
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			log.Fatalf("close drop migration: source=%v database=%v", sourceErr, dbErr)
		}

		m, err = newMigrate(*path, cfg.DB.URL())
		if err != nil {
			log.Fatalf("create fresh migration: %v", err)
		}
		defer m.Close()

		if err := migrateUp(m); err != nil {
			log.Fatalf("run fresh migration: %v", err)
		}
		log.Println("migration fresh completed")
	default:
		log.Fatalf("unsupported migration direction %q, supported directions: up, fresh", *direction)
	}
}

func newMigrate(path, databaseURL string) (*migrate.Migrate, error) {
	return migrate.New("file://"+path, databaseURL)
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

func isMissingSchemaMigrations(err error) bool {
	return strings.Contains(err.Error(), `relation "public.schema_migrations" does not exist`)
}
