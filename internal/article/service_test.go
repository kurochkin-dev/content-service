package article

import (
	"errors"
	"testing"
)

type mockRepository struct {
	articles map[uint]*Article
	nextID   uint
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		articles: make(map[uint]*Article),
		nextID:   1,
	}
}

func (m *mockRepository) Create(article *Article) error {
	article.ID = m.nextID
	m.nextID++
	m.articles[article.ID] = article
	return nil
}

func (m *mockRepository) GetByID(id uint) (*Article, error) {
	article, ok := m.articles[id]
	if !ok {
		return nil, ErrNotFound
	}
	return article, nil
}

func (m *mockRepository) GetAll(page, limit int) ([]Article, int64, error) {
	allArticles := make([]Article, 0, len(m.articles))
	for _, article := range m.articles {
		allArticles = append(allArticles, *article)
	}

	total := int64(len(allArticles))

	offset := (page - 1) * limit

	if offset >= len(allArticles) {
		return []Article{}, total, nil
	}

	end := offset + limit
	if end > len(allArticles) {
		end = len(allArticles)
	}

	return allArticles[offset:end], total, nil
}

func (m *mockRepository) Update(id uint, updates map[string]interface{}) error {
	article, ok := m.articles[id]
	if !ok {
		return ErrNotFound
	}
	if title, ok := updates["title"].(string); ok {
		article.Title = title
	}
	if content, ok := updates["content"].(string); ok {
		article.Content = content
	}
	return nil
}

func (m *mockRepository) Delete(id uint) error {
	if _, ok := m.articles[id]; !ok {
		return ErrNotFound
	}
	delete(m.articles, id)
	return nil
}

func TestCreateArticle(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	tests := []struct {
		name      string
		userID    uint
		title     string
		content   string
		wantError bool
	}{
		{
			name:      "Valid article",
			userID:    1,
			title:     "Test Article",
			content:   "Test Content",
			wantError: false,
		},
		{
			name:      "Empty title",
			userID:    1,
			title:     "",
			content:   "Test Content",
			wantError: true,
		},
		{
			name:      "Empty content",
			userID:    1,
			title:     "Test Article",
			content:   "",
			wantError: true,
		},
		{
			name:      "Zero user ID",
			userID:    0,
			title:     "Test Article",
			content:   "Test Content",
			wantError: true,
		},
		{
			name:      "Title too long",
			userID:    1,
			title:     string(make([]byte, 256)),
			content:   "Test Content",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article, err := svc.CreateArticle(tt.userID, tt.title, tt.content)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if article == nil {
					t.Errorf("Expected article but got nil")
				}
				if article.Title != tt.title {
					t.Errorf("Expected title %q, got %q", tt.title, article.Title)
				}
			}
		})
	}
}

func TestGetArticleByID(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	article, err := svc.CreateArticle(1, "Test", "Content")
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	tests := []struct {
		name      string
		id        uint
		wantError bool
	}{
		{
			name:      "Existing article",
			id:        article.ID,
			wantError: false,
		},
		{
			name:      "Non-existing article",
			id:        999,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := svc.GetArticleByID(tt.id)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("Expected ErrNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if found == nil {
					t.Errorf("Expected article but got nil")
				}
			}
		})
	}
}

func TestUpdateArticle(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	article, err := svc.CreateArticle(1, "Original Title", "Original Content")
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	newTitle := "Updated Title"
	newContent := "Updated Content"

	tests := []struct {
		name      string
		userID    uint
		id        uint
		title     *string
		content   *string
		wantError bool
	}{
		{
			name:      "Update title",
			userID:    1,
			id:        article.ID,
			title:     &newTitle,
			content:   nil,
			wantError: false,
		},
		{
			name:      "Update content",
			userID:    1,
			id:        article.ID,
			title:     nil,
			content:   &newContent,
			wantError: false,
		},
		{
			name:      "Wrong user",
			userID:    2,
			id:        article.ID,
			title:     &newTitle,
			content:   nil,
			wantError: true,
		},
		{
			name:      "Non-existing article",
			userID:    1,
			id:        999,
			title:     &newTitle,
			content:   nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := svc.UpdateArticle(tt.userID, tt.id, tt.title, tt.content)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if updated == nil {
					t.Errorf("Expected updated article but got nil")
				}
			}
		})
	}
}

func TestDeleteArticle(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	article, err := svc.CreateArticle(1, "Test", "Content")
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	tests := []struct {
		name      string
		userID    uint
		id        uint
		wantError bool
	}{
		{
			name:      "Wrong user",
			userID:    2,
			id:        article.ID,
			wantError: true,
		},
		{
			name:      "Correct user",
			userID:    1,
			id:        article.ID,
			wantError: false,
		},
		{
			name:      "Already deleted",
			userID:    1,
			id:        article.ID,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteArticle(tt.userID, tt.id)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetAllArticles(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	for i := 1; i <= 5; i++ {
		_, err := svc.CreateArticle(1, "Article", "Content")
		if err != nil {
			t.Fatalf("Failed to create test article: %v", err)
		}
	}

	tests := []struct {
		name      string
		page      int
		limit     int
		wantCount int
	}{
		{
			name:      "Default pagination",
			page:      1,
			limit:     10,
			wantCount: 5,
		},
		{
			name:      "Invalid page",
			page:      0,
			limit:     10,
			wantCount: 5,
		},
		{
			name:      "Invalid limit",
			page:      1,
			limit:     0,
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles, total, err := svc.GetAllArticles(tt.page, tt.limit)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(articles) != tt.wantCount {
				t.Errorf("Expected %d articles, got %d", tt.wantCount, len(articles))
			}
			if int(total) != tt.wantCount {
				t.Errorf("Expected total %d, got %d", tt.wantCount, total)
			}
		})
	}
}
