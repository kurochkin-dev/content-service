package article

import (
	"errors"
	"net/http"
	"strconv"

	"content-service/internal/shared/middleware"
	"content-service/internal/shared/validation"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type CreateArticleRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"required,min=1"`
}

type UpdateArticleRequest struct {
	Title   *string `json:"title" validate:"omitempty,min=1,max=255"`
	Content *string `json:"content" validate:"omitempty,min=1"`
}

func getID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

var errorToStatus = map[error]int{
	ErrNotFound:   http.StatusNotFound,
	ErrForbidden:  http.StatusForbidden,
	ErrValidation: http.StatusBadRequest,
}

func (handler *Handler) handleError(c *gin.Context, err error) {
	for target, status := range errorToStatus {
		if errors.Is(err, target) {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
	}

	log.Error().Err(err).Msg("Internal error")
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func (handler *Handler) CreateArticle(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := validation.NormalizeValidationErrors(err, req)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	article, err := handler.service.CreateArticle(userID, req.Title, req.Content)
	if err != nil {
		handler.handleError(c, err)
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
		handler.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

func (handler *Handler) GetAllArticles(c *gin.Context) {
	page := DefaultPage
	limit := DefaultLimit

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	articles, total, err := handler.service.GetAllArticles(page, limit)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
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

	var updateReq UpdateArticleRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		validationErrors := validation.NormalizeValidationErrors(err, updateReq)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	if updateReq.Title == nil && updateReq.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field (title or content) must be provided"})
		return
	}

	updatedArticle, err := handler.service.UpdateArticle(userID, id, updateReq.Title, updateReq.Content)
	if err != nil {
		handler.handleError(c, err)
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

	if err := handler.service.DeleteArticle(userID, id); err != nil {
		handler.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
