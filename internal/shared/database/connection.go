package database

import (
	"fmt"

	"content-service/internal/shared/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	var logLevel logger.LogLevel
	if cfg.IsProduction() {
		logLevel = logger.Error
	} else {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DB.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
