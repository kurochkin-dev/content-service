package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	DB          DBConfig
	App         AppConfig
	JWT         JWTConfig
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type AppConfig struct {
	Port    int
	GinMode string
}

type JWTConfig struct {
	Secret string
}

func LoadConfig() (*Config, error) {
	if os.Getenv("ENVIRONMENT") != "production" {
		_ = godotenv.Load()
	}

	env := strings.ToLower(getEnv("ENVIRONMENT", "development"))
	jwtSecret := getEnv("JWT_SECRET", "")
	ginMode := getEnv("GIN_MODE", "")

	if env != "production" && len(jwtSecret) < 32 {
		jwtSecret = "dev-secret-key-min-32-chars------"
	}

	if ginMode == "" {
		if env == "production" {
			ginMode = "release"
		} else {
			ginMode = "debug"
		}
	}

	cfg := &Config{
		Environment: env,
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "content_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
			ConnMaxIdleTime: time.Duration(getEnvInt("DB_CONN_MAX_IDLE_TIME_MIN", 2)) * time.Minute,
		},
		App: AppConfig{
			Port:    getEnvInt("PORT", 8080),
			GinMode: ginMode,
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"test":        true,
		"production":  true,
	}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid ENVIRONMENT: must be one of: development, staging, test, production")
	}

	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("invalid PORT: must be 1..65535")
	}

	if c.DB.Host == "" {
		return fmt.Errorf("invalid DB_HOST: cannot be empty")
	}
	if c.DB.User == "" {
		return fmt.Errorf("invalid DB_USER: cannot be empty")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("invalid DB_NAME: cannot be empty")
	}
	if c.DB.Port < 1 || c.DB.Port > 65535 {
		return fmt.Errorf("invalid DB_PORT: must be 1..65535")
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[c.DB.SSLMode] {
		return fmt.Errorf("invalid DB_SSLMODE: must be one of: disable, require, verify-ca, verify-full")
	}

	if c.Environment == "production" {
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("invalid JWT_SECRET: must be >= 32 chars in production")
		}
	} else {
		if c.JWT.Secret == "" {
			return fmt.Errorf("invalid JWT_SECRET: cannot be empty")
		}
	}

	if c.Environment == "production" && c.App.GinMode != "release" {
		return fmt.Errorf("invalid GIN_MODE: must be 'release' in production")
	}

	return nil
}

func (c *Config) GetDSN() string {
	escapedPassword := url.QueryEscape(c.DB.Password)
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		escapedPassword,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func (c *Config) GetSafeDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return defaultVal
}
