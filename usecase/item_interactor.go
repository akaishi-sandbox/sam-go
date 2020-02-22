package usecase

import (
	elastic "github.com/olivere/elastic/v7"
)

type ItemInteractor struct {
	ItemRepository ItemRepository
}

func (interactor *ItemInteractor) Search(q map[string]string) (interface{}, error) {
	searchResult, err := interactor.ItemRepository.Search(q)
	if err != nil {
		return nil, err
	}
	return struct {
		Total int64                `json:"total"`
		Items []*elastic.SearchHit `json:"items"`
	}{
		Total: searchResult.TotalHits(),
		Items: searchResult.Hits.Hits,
	}, nil
}
