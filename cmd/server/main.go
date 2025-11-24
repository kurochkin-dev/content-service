package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"content-service/internal/article"
	"content-service/internal/shared/config"
	"content-service/internal/shared/database"
	"content-service/internal/shared/logging"
	"content-service/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	logging.InitLogger(cfg.Environment)
	log.Info().Str("environment", cfg.Environment).Msg("Starting content-service")

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	autoMigrate := os.Getenv("AUTO_MIGRATE")
	if autoMigrate == "" {
		autoMigrate = "true"
		if cfg.IsProduction() {
			autoMigrate = "false"
		}
	}

	if autoMigrate == "true" {
		if err := db.AutoMigrate(&article.Article{}); err != nil {
			log.Fatal().Err(err).Msg("Failed to run migrations")
		}
		log.Info().Msg("Database AutoMigrate completed")
	} else {
		log.Info().Msg("AutoMigrate disabled - use './migrate' command for schema changes")
	}

	gin.SetMode(cfg.App.GinMode)

	articleRepo := article.NewRepository(db)
	articleService := article.NewService(articleRepo)
	articleHandler := article.NewHandler(articleService)

	router := gin.Default()

	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.CORSMiddleware(cfg))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "content-service",
		})
	})

	api := router.Group("/api")
	{
		articles := api.Group("/articles")
		{
			articles.POST("", middleware.JWTAuthMiddleware(cfg), articleHandler.CreateArticle)
			articles.GET("", articleHandler.GetAllArticles)
			articles.GET("/:id", articleHandler.GetArticleByID)
			articles.PUT("/:id", middleware.JWTAuthMiddleware(cfg), articleHandler.UpdateArticle)
			articles.DELETE("/:id", middleware.JWTAuthMiddleware(cfg), articleHandler.DeleteArticle)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.App.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("address", addr).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		} else {
			log.Info().Msg("Database connection closed")
		}
	}

	log.Info().Msg("Server exited gracefully")
}
