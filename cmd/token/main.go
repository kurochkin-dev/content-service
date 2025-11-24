package main

import (
	"flag"
	"fmt"

	"content-service/internal/shared/config"
	"content-service/internal/shared/logging"
	"content-service/internal/shared/middleware"

	"github.com/rs/zerolog/log"
)

func main() {
	var userID = flag.Uint("user-id", 1, "User ID for the token")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	logging.InitLogger(cfg.Environment)

	if *userID == 0 {
		log.Fatal().Msg("user-id cannot be 0")
	}

	if cfg.JWT.Secret == "" {
		log.Fatal().Msg("JWT_SECRET is not set")
	}

	token, err := middleware.CreateTestToken(*userID, cfg.JWT.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token")
	}

	fmt.Println(token)
}
