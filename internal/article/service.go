package article

import (
	"fmt"
)

type Service interface {
	CreateArticle(userID uint, title, content string) (*Article, error)
	GetArticleByID(id uint) (*Article, error)
	GetAllArticles() ([]Article, error)
	UpdateArticle(id uint, title, content *string) (*Article, error)
	DeleteArticle(id uint) error
}

type articleService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &articleService{repo: repo}
}

func (svc *articleService) CreateArticle(userID uint, title, content string) (*Article, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id cannot be empty")
	}
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	article := &Article{
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	if err := svc.repo.Create(article); err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
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

func (svc *articleService) GetAllArticles() ([]Article, error) {
	articles, err := svc.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	return articles, nil
}

func (svc *articleService) UpdateArticle(id uint, title, content *string) (*Article, error) {
	article, err := svc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if title != nil {
		if *title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		if len(*title) > 255 {
			return nil, fmt.Errorf("title cannot exceed 255 characters")
		}
		updates["title"] = *title
		article.Title = *title
	}
	if content != nil {
		if *content == "" {
			return nil, fmt.Errorf("content cannot be empty")
		}
		updates["content"] = *content
		article.Content = *content
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	if err := svc.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return article, nil
}

func (svc *articleService) DeleteArticle(id uint) error {
	if err := svc.repo.Delete(id); err != nil {
		return err
	}
	return nil
}
