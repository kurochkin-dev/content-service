package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"content-service/internal/shared/config"
	"content-service/internal/shared/middleware"
)

func main() {
	var userID = flag.Uint("user-id", 1, "User ID for the token")
	flag.Parse()

	if *userID == 0 {
		log.Fatal("user-id cannot be 0")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.JWT.Secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	token, err := middleware.CreateTestToken(uint(*userID), cfg.JWT.Secret)
	if err != nil {
		log.Fatalf("failed to create token: %v", err)
	}

	fmt.Println(token)
	os.Exit(0)
}
