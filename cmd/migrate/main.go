package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"content-service/internal/shared/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var command = flag.String("command", "up", "migration command: up, down, version, force")
	var versionFlag = flag.Int("version", 0, "version for force command")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	migrationsPath := "file://./migrations"
	migrationsAbsPath, err := filepath.Abs("./migrations")
	if err != nil {
		log.Fatalf("failed to get absolute path for migrations: %v", err)
	}
	migrationsPath = fmt.Sprintf("file://%s", migrationsAbsPath)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	migrator, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	defer migrator.Close()

	switch *command {
	case "up":
		if err := migrator.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to apply")
				return
			}
			log.Fatalf("failed to run migrations up: %v", err)
		}
		version, dirty, _ := migrator.Version()
		log.Printf("Migrations applied successfully. Current version: %d, dirty: %v", version, dirty)

	case "down":
		if err := migrator.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to rollback")
				return
			}
			log.Fatalf("failed to run migrations down: %v", err)
		}
		version, dirty, _ := migrator.Version()
		log.Printf("Migrations rolled back successfully. Current version: %d, dirty: %v", version, dirty)

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Println("No migrations applied yet")
				return
			}
			log.Fatalf("failed to get version: %v", err)
		}
		log.Printf("Current migration version: %d, dirty: %v", version, dirty)

	case "force":
		if *versionFlag == 0 {
			log.Fatal("version flag is required for force command")
		}
		if err := migrator.Force(*versionFlag); err != nil {
			log.Fatalf("failed to force version: %v", err)
		}
		log.Printf("Force version set to %d", *versionFlag)

	default:
		log.Fatalf("unknown command: %s. Use: up, down, version, force", *command)
	}
}
