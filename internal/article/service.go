package article

import (
	"fmt"
)

type Service interface {
	CreateArticle(userID uint, title, content string) (*Article, error)
	GetArticleByID(id uint) (*Article, error)
	GetAllArticles(page, limit int) ([]Article, int64, error)
	UpdateArticle(userID, id uint, title, content *string) (*Article, error)
	DeleteArticle(userID, id uint) error
}

type articleService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &articleService{repo: repo}
}

func (svc *articleService) CreateArticle(userID uint, title, content string) (*Article, error) {
	if userID == 0 {
		return nil, fmt.Errorf("%w: user_id cannot be empty", ErrValidation)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(title) > MaxTitleLength {
		return nil, fmt.Errorf("%w: title cannot exceed %d characters", ErrValidation, MaxTitleLength)
	}
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrValidation)
	}

	article := &Article{
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	if err := svc.repo.Create(article); err != nil {
		return nil, fmt.Errorf("%w: failed to create article: %w", ErrInternal, err)
	}

	return article, nil
}

func (svc *articleService) GetArticleByID(id uint) (*Article, error) {
	article, err := svc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (svc *articleService) GetAllArticles(page, limit int) ([]Article, int64, error) {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	articles, total, err := svc.repo.GetAll(page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: failed to get articles: %w", ErrInternal, err)
	}
	return articles, total, nil
}

func (svc *articleService) UpdateArticle(userID, id uint, title, content *string) (*Article, error) {
	article, err := svc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if article.UserID != userID {
		return nil, ErrForbidden
	}

	updates := make(map[string]interface{})

	if title != nil {
		if *title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", ErrValidation)
		}
		if len(*title) > MaxTitleLength {
			return nil, fmt.Errorf("%w: title cannot exceed %d characters", ErrValidation, MaxTitleLength)
		}
		updates["title"] = *title
		article.Title = *title
	}

	if content != nil {
		if *content == "" {
			return nil, fmt.Errorf("%w: content cannot be empty", ErrValidation)
		}
		updates["content"] = *content
		article.Content = *content
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("%w: no fields to update", ErrValidation)
	}

	if err := svc.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("%w: failed to update article: %w", ErrInternal, err)
	}

	return article, nil
}

func (svc *articleService) DeleteArticle(userID, id uint) error {
	article, err := svc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if article.UserID != userID {
		return ErrForbidden
	}

	if err := svc.repo.Delete(id); err != nil {
		return fmt.Errorf("%w: failed to delete article: %w", ErrInternal, err)
	}

	return nil
}
