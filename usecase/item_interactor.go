package usecase

import (
	elastic "github.com/olivere/elastic/v7"
)

// ItemInteractor struct
type ItemInteractor struct {
	ItemRepository ItemRepository
}

// Search function
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

// Recommend function
func (interactor *ItemInteractor) Recommend(q map[string]string) (interface{}, error) {
	searchResult, err := interactor.ItemRepository.Recommend(q)
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

// Classification function
func (interactor *ItemInteractor) Classification(q map[string]string) (interface{}, error) {
	searchResult, err := interactor.ItemRepository.Classification(q)
	if err != nil {
		return nil, err
	}
	return struct {
		Total int64                `json:"total"`
		Hits  []*elastic.SearchHit `json:"hits"`
	}{
		Total: searchResult.TotalHits(),
		Hits:  searchResult.Hits.Hits,
	}, nil
}

// AccessInfo function
func (interactor *ItemInteractor) AccessInfo(q map[string]string) (interface{}, error) {
	updateItem, err := interactor.ItemRepository.AccessInfo(q)
	if err != nil {
		return nil, err
	}
	return updateItem, nil
}
