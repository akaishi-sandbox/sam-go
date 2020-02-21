package classificationinfo

import (
	"strconv"
	"strings"

	searchitems "github.com/akaishi-sandbox/sam-go/internal/search-items"
	"github.com/akaishi-sandbox/sam-go/pkg"
	elastic "github.com/olivere/elastic/v7"
)

// CreateRecommendItems elastic search query
func CreateRecommendItems(item searchitems.Item, q map[string]string) *pkg.SearchQuery {
	query := elastic.NewBoolQuery()
	if itemID, ok := q["item_id"]; ok {
		query = query.MustNot(elastic.NewTermQuery("item_id", itemID))
	}
	if brand, ok := q["brand"]; ok {
		query = query.Filter(pkg.NewTermsString("brand", strings.Split(brand, ",")))
	}
	query = query.Filter(pkg.NewTermsString("gender", strings.Split(item.Gender, ",")))
	query = query.Filter(pkg.NewTermsString("category", strings.Split(item.Category, ",")))

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

	return &pkg.SearchQuery{
		Index: "items",
		Query: query,
		From:  from,
		Size:  size,
	}
}
