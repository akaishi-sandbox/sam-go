package usecase

import (
	elastic "github.com/olivere/elastic/v7"
)

type ItemRepository interface {
	Search(q map[string]string) (*elastic.SearchResult, error)
}
