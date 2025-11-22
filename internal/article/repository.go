package article

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	Create(article *Article) error
	GetByID(id uint) (*Article, error)
	GetAll(page, limit int) ([]Article, int64, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
}

type articleRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &articleRepository{db: db}
}

func (repo *articleRepository) Create(article *Article) error {
	if err := repo.db.Create(article).Error; err != nil {
		return fmt.Errorf("repo: failed to create article: %w", err)
	}
	return nil
}

func (repo *articleRepository) GetByID(id uint) (*Article, error) {
	var article Article
	err := repo.db.First(&article, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repo: failed to get article by id %d: %w", id, err)
	}
	return &article, nil
}

func (repo *articleRepository) GetAll(page, limit int) ([]Article, int64, error) {
	var articles []Article
	var total int64

	if err := repo.db.Model(&Article{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repo: failed to count articles: %w", err)
	}

	offset := (page - 1) * limit

	err := repo.db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&articles).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repo: failed to get articles: %w", err)
	}

	return articles, total, nil
}

func (repo *articleRepository) Update(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("repo: no fields to update")
	}

	updateResult := repo.db.Model(&Article{}).Where("id = ?", id).Updates(updates)
	if updateResult.Error != nil {
		return fmt.Errorf("repo: failed to update article %d: %w", id, updateResult.Error)
	}
	if updateResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *articleRepository) Delete(id uint) error {
	deleteResult := repo.db.Delete(&Article{}, id)
	if deleteResult.Error != nil {
		return fmt.Errorf("repo: failed to delete article %d: %w", id, deleteResult.Error)
	}
	if deleteResult.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
