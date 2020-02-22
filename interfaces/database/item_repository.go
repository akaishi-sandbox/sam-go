package database

import (
	"strconv"
	"strings"

	"github.com/akaishi-sandbox/sam-go/infrastructure"
	"github.com/akaishi-sandbox/sam-go/pkg"
	elastic "github.com/olivere/elastic/v7"
)

type ItemRepository struct {
	ElasticHandler *infrastructure.ElasticHandler
}

func createSearchQuery(q map[string]string) *infrastructure.ElasticQuery {
	query := elastic.NewBoolQuery()
	if itemID, ok := q["item_id"]; ok && len(itemID) > 0 {
		query = query.Filter(pkg.NewTermsString("item_id", strings.Split(itemID, ",")))
	}
	if gender, ok := q["gender"]; ok && len(gender) > 0 {
		query = query.Filter(pkg.NewTermsString("gender", strings.Split(gender, ",")))
	}
	if brand, ok := q["brand"]; ok && len(brand) > 0 {
		query = query.Filter(pkg.NewTermsString("brand", strings.Split(brand, ",")))
	}
	if category, ok := q["category"]; ok && len(category) > 0 {
		query = query.Filter(pkg.NewTermsString("category", strings.Split(category, ",")))
	}
	if discountFlag, ok := q["discount_flag"]; ok && len(discountFlag) > 0 {
		query = query.Filter(pkg.NewTermsString("discount_flag", strings.Split(discountFlag, ",")))
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

func (repo *ItemRepository) Search(q map[string]string) (*elastic.SearchResult, error) {
	query := createSearchQuery(q)
	return repo.ElasticHandler.Search(query)
}
