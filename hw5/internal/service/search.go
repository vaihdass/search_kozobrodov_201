package service

import (
	"github.com/vaihdass/search_kozobrodov_201/search/internal/search"
)

// Result - DTO результата поиска.
type Result struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Snippet    string   `json:"snippet"`
	Highlights []string `json:"highlights"`
	Score      float64  `json:"score"`
}

// SearchService предоставляет поиск по документам.
type SearchService struct {
	engine *search.Engine
}

// New создает сервис поиска, загружая данные из dataDir.
func New(dataDir string) (*SearchService, error) {
	engine, err := search.New(dataDir)
	if err != nil {
		return nil, err
	}
	return &SearchService{engine: engine}, nil
}

// Search выполняет поиск и возвращает результаты.
func (s *SearchService) Search(query string, limit int) []Result {
	raw := s.engine.Search(query, limit)

	results := make([]Result, len(raw))
	for i, r := range raw {
		results[i] = Result{
			Title:      r.Title,
			URL:        r.URL,
			Snippet:    r.Snippet,
			Highlights: r.Highlights,
			Score:      r.Score,
		}
	}

	return results
}
