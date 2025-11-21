package article

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"content-service/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required,min=1"`
}

type UpdateArticleRequest struct {
	Title   *string `json:"title" binding:"required_without=Content,min=1,max=255"`
	Content *string `json:"content" binding:"required_without=Title,min=1"`
}

func getID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func normalizeValidationError(err error, req interface{}) []string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []string{"validation failed"}
	}

	var errorsList []string
	reqType := reflect.TypeOf(req)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	for _, fieldErr := range validationErrors {
		jsonName := fieldErr.Field()

		if field, found := reqType.FieldByName(fieldErr.StructField()); found {
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				if commaIndex := strings.Index(jsonTag, ","); commaIndex > 0 {
					jsonName = jsonTag[:commaIndex]
				} else {
					jsonName = jsonTag
				}
			}
		}

		var message string
		switch fieldErr.Tag() {
		case "required", "required_without":
			message = jsonName + " is required"
		case "min":
			message = jsonName + " is too short"
		case "max":
			message = jsonName + " is too long"
		default:
			message = jsonName + " validation failed"
		}
		errorsList = append(errorsList, message)
	}

	return errorsList
}

func (handler *Handler) CreateArticle(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": normalizeValidationError(err, req)})
		return
	}

	article, err := handler.service.CreateArticle(userID, req.Title, req.Content)
	if err != nil {
		log.Printf("error creating article: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, article)
}

func (handler *Handler) GetArticleByID(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	article, err := handler.service.GetArticleByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		log.Printf("error getting article %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get article"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (handler *Handler) GetAllArticles(c *gin.Context) {
	articles, err := handler.service.GetAllArticles()
	if err != nil {
		log.Printf("error getting articles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

func (handler *Handler) UpdateArticle(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	id, err := getID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	article, err := handler.service.GetArticleByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		log.Printf("error getting article %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get article"})
		return
	}

	if article.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own articles"})
		return
	}

	var updateReq UpdateArticleRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": normalizeValidationError(err, updateReq)})
		return
	}

	if updateReq.Title == nil && updateReq.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	updatedArticle, err := handler.service.UpdateArticle(id, updateReq.Title, updateReq.Content)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		log.Printf("error updating article %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
		return
	}

	c.JSON(http.StatusOK, updatedArticle)
}

func (handler *Handler) DeleteArticle(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	id, err := getID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	article, err := handler.service.GetArticleByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		log.Printf("error getting article %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get article"})
		return
	}

	if article.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own articles"})
		return
	}

	if err := handler.service.DeleteArticle(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		log.Printf("error deleting article %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete article"})
		return
	}

	c.Status(http.StatusNoContent)
}
