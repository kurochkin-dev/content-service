package main

import (
	"fmt"
	"log"

	"content-service/internal/article"
	"content-service/internal/shared/config"
	"content-service/internal/shared/database"
	"content-service/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&article.Article{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	gin.SetMode(cfg.App.GinMode)

	articleRepo := article.NewRepository(db)
	articleService := article.NewService(articleRepo)
	articleHandler := article.NewHandler(articleService)

	router := gin.Default()

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
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
