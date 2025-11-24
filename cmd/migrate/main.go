package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"content-service/internal/shared/config"
	"content-service/internal/shared/logging"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func main() {
	var command = flag.String("command", "up", "migration command: up, down, version, force")
	var versionFlag = flag.Int("version", 0, "version for force command")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	logging.InitLogger(cfg.Environment)

	migrationsPath := "file://./migrations"
	migrationsAbsPath, err := filepath.Abs("./migrations")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get absolute path for migrations")
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
		log.Fatal().Err(err).Msg("Failed to create migrator")
	}
	defer func() {
		sourceErr, dbErr := migrator.Close()
		if sourceErr != nil {
			log.Error().Err(sourceErr).Msg("Failed to close migrator source")
		}
		if dbErr != nil {
			log.Error().Err(dbErr).Msg("Failed to close migrator database")
		}
	}()

	switch *command {
	case "up":
		if err := migrator.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info().Msg("No migrations to apply")
				return
			}
			log.Fatal().Err(err).Msg("Failed to run migrations up")
		}
		version, dirty, _ := migrator.Version()
		log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrations applied successfully")

	case "down":
		if err := migrator.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info().Msg("No migrations to rollback")
				return
			}
			log.Fatal().Err(err).Msg("Failed to run migrations down")
		}
		version, dirty, _ := migrator.Version()
		log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrations rolled back successfully")

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				log.Info().Msg("No migrations applied yet")
				return
			}
			log.Fatal().Err(err).Msg("Failed to get version")
		}
		log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Current migration version")

	case "force":
		if *versionFlag == 0 {
			log.Fatal().Msg("version flag is required for force command")
		}
		if err := migrator.Force(*versionFlag); err != nil {
			log.Fatal().Err(err).Msg("Failed to force version")
		}
		log.Info().Int("version", *versionFlag).Msg("Force version set")

	default:
		log.Fatal().Str("command", *command).Msg("Unknown command. Use: up, down, version, force")
	}
}
