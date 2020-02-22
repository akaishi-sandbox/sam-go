package usecase

import (
	"github.com/akaishi-sandbox/sam-go/domain"
	elastic "github.com/olivere/elastic/v7"
)

// ItemRepository interface
type ItemRepository interface {
	Search(q map[string]string) (*elastic.SearchResult, error)
	Recommend(q map[string]string) (*elastic.SearchResult, error)
	Classification(q map[string]string) (*elastic.SearchResult, error)
	AccessInfo(q map[string]string) (*domain.Item, error)
}
