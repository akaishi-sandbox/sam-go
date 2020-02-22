package database

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/akaishi-sandbox/sam-go/domain"
	"github.com/akaishi-sandbox/sam-go/infrastructure"
	elastic "github.com/olivere/elastic/v7"
)

// ItemRepository struct
type ItemRepository struct {
	ElasticHandler *infrastructure.ElasticHandler
}

func newTermsString(name string, input []string) *elastic.TermsQuery {
	values := make([]interface{}, len(input))
	for i, s := range input {
		values[i] = s
	}
	return elastic.NewTermsQuery(name, values...)
}

func createSearchQuery(q map[string]string) *infrastructure.ElasticQuery {
	query := elastic.NewBoolQuery()
	if itemID, ok := q["item_id"]; ok && len(itemID) > 0 {
		query = query.Filter(newTermsString("item_id", strings.Split(itemID, ",")))
	}
	if gender, ok := q["gender"]; ok && len(gender) > 0 {
		query = query.Filter(newTermsString("gender", strings.Split(gender, ",")))
	}
	if brand, ok := q["brand"]; ok && len(brand) > 0 {
		query = query.Filter(newTermsString("brand", strings.Split(brand, ",")))
	}
	if category, ok := q["category"]; ok && len(category) > 0 {
		query = query.Filter(newTermsString("category", strings.Split(category, ",")))
	}
	if discountFlag, ok := q["discount_flag"]; ok && len(discountFlag) > 0 {
		query = query.Filter(newTermsString("discount_flag", strings.Split(discountFlag, ",")))
	}
	if minPrice, ok := q["min_price"]; ok {
		if price, err := strconv.Atoi(minPrice); err == nil {
			query = query.Filter(elastic.NewRangeQuery("lowest_price").Gte(price))
		}
	}
	if maxPrice, ok := q["max_price"]; ok {
		if price, err := strconv.Atoi(maxPrice); err == nil {
			query = query.Filter(elastic.NewRangeQuery("lowest_price").Lte(price))
		}
	}
	if keywords, ok := q["keywords"]; ok && len(keywords) > 0 {
		// 全角スペースを半角スペースにした後、半角スペースで分解する
		for _, keyword := range strings.Split(strings.NewReplacer("　", " ").Replace(keywords), " ") {
			// キーワードの中にスラッシュがある場合ORとして評価する
			if strings.Index(keyword, "/") != -1 {
				q := elastic.NewBoolQuery()
				for _, parseWord := range strings.Split(keyword, "/") {
					q.Should(elastic.NewMatchPhraseQuery("search_text", parseWord))
				}
				query = query.Must(q)
			} else {
				// スラッシュを含んでいない場合はANDとして評価する
				query = query.Must(elastic.NewMatchPhraseQuery("search_text", keyword))
			}
		}
	}

	if excludeExpired, ok := q["exclude_expired"]; !ok && excludeExpired == "1" {
		query = query.Filter(elastic.NewTermsQuery("release_flag", 0, 1))
	}

	from := 0
	if offset, ok := q["offset"]; ok {
		if v, err := strconv.Atoi(offset); err == nil {
			from = v
		}
	}
	size := 36
	if limit, ok := q["limit"]; ok {
		if v, err := strconv.Atoi(limit); err == nil {
			size = v
		}
	}

	sort := elastic.SortInfo{Field: "updated_at", Ascending: false}

	if order, ok := q["order"]; ok {
		switch order {
		case "new":
		case "min-max":
			sort.Field = "lowest_price"
			sort.Ascending = true
		case "max-max":
			sort.Field = "lowest_price"
			sort.Ascending = false
		}
	}

	return &infrastructure.ElasticQuery{
		Index:    "items",
		Query:    query,
		SortInfo: sort,
		From:     from,
		Size:     size,
	}
}

func createRecommendItems(item domain.Item, q map[string]string) *infrastructure.ElasticQuery {
	query := elastic.NewBoolQuery()
	if itemID, ok := q["item_id"]; ok {
		query = query.MustNot(elastic.NewTermQuery("item_id", itemID))
	}
	if brand, ok := q["brand"]; ok {
		query = query.Filter(newTermsString("brand", strings.Split(brand, ",")))
	}
	query = query.Filter(newTermsString("gender", strings.Split(item.Gender, ",")))
	query = query.Filter(newTermsString("category", strings.Split(item.Category, ",")))

	from := 0
	if offset, ok := q["offset"]; ok {
		if v, err := strconv.Atoi(offset); err == nil {
			from = v
		}
	}
	size := 36
	if limit, ok := q["limit"]; ok {
		if v, err := strconv.Atoi(limit); err == nil {
			size = v
		}
	}

	return &infrastructure.ElasticQuery{
		Index: "items",
		Query: query,
		From:  from,
		Size:  size,
	}
}

func createClassificationQuery(q map[string]string) (*infrastructure.ElasticQuery, error) {
	query := elastic.NewBoolQuery()
	if gender, ok := q["gender"]; ok {
		query = query.Filter(newTermsString("gender", strings.Split(gender, ",")))
	}
	if title, ok := q["title"]; ok {
		query = query.Filter(newTermsString("title", strings.Split(title, ",")))
	}
	from := 0
	if offset, ok := q["offset"]; ok {
		if v, err := strconv.Atoi(offset); err == nil {
			from = v
		}
	}
	size := 100
	if limit, ok := q["limit"]; ok {
		if v, err := strconv.Atoi(limit); err == nil {
			size = v
		}
	}

	sort := elastic.SortInfo{Field: "sort_no", Ascending: true}

	index, ok := q["index"]
	if !ok {
		return nil, fmt.Errorf("parameter not found")
	}
	switch index {
	case "categories", "brands":
		return &infrastructure.ElasticQuery{
			Index:    index,
			Query:    query,
			SortInfo: sort,
			From:     from,
			Size:     size,
		}, nil
	default:
		return nil, fmt.Errorf("not supported index")
	}

}

// Search function
func (repo *ItemRepository) Search(q map[string]string) (*elastic.SearchResult, error) {
	return repo.ElasticHandler.Search(createSearchQuery(q))
}

// Recommend function
func (repo *ItemRepository) Recommend(q map[string]string) (*elastic.SearchResult, error) {
	query := elastic.NewBoolQuery()
	itemID, ok := q["item_id"]
	if !ok {
		return nil, fmt.Errorf("parameter not found")
	}

	query = query.Filter(elastic.NewTermQuery("item_id", itemID))

	searchResult, err := repo.ElasticHandler.Search(&infrastructure.ElasticQuery{
		Index: "items",
		Query: query,
		From:  0,
		Size:  1,
	})
	if err != nil {
		return nil, err
	}

	var item domain.Item
	if err := json.Unmarshal(searchResult.Hits.Hits[0].Source, &item); err != nil {
		return nil, err
	}

	return repo.ElasticHandler.Search(createRecommendItems(item, q))
}

// Classification function
func (repo *ItemRepository) Classification(q map[string]string) (*elastic.SearchResult, error) {
	query, err := createClassificationQuery(q)
	if err != nil {
		return nil, err
	}
	return repo.ElasticHandler.Search(query)
}

// AccessInfo function
func (repo *ItemRepository) AccessInfo(q map[string]string) (*domain.Item, error) {
	query := elastic.NewBoolQuery()
	itemID, ok := q["item_id"]
	if !ok {
		return nil, fmt.Errorf("parameter not found")
	}

	query = query.Filter(elastic.NewTermQuery("item_id", itemID))

	searchResult, err := repo.ElasticHandler.Search(&infrastructure.ElasticQuery{
		Index: "items",
		Query: query,
		From:  0,
		Size:  100,
	})
	if err != nil {
		return nil, err
	}

	// 更新元の商品はIDを元に検索しているので複数個存在する場合がある、そのためアクセス回数の最も大きい値を更新元の数字として取得する
	updateItem := &domain.Item{
		AccessCounter:  0,
		LastAccessedAt: time.Now(),
	}
	var iType domain.Item
	for _, item := range searchResult.Each(reflect.TypeOf(iType)) {
		if i, ok := item.(domain.Item); ok {
			if i.AccessCounter > updateItem.AccessCounter {
				updateItem.AccessCounter = i.AccessCounter
			}
		}
	}
	updateItem.AccessCounter++

	for _, hit := range searchResult.Hits.Hits {
		if _, err := repo.ElasticHandler.Update(hit, updateItem); err != nil {
			return nil, err
		}
	}

	return updateItem, nil
}
