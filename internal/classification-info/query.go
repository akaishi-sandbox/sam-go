package classificationinfo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/akaishi-sandbox/sam-go/pkg"
	elastic "github.com/olivere/elastic/v7"
)

// CreateSearchQuery elastic search query
func CreateSearchQuery(q map[string]string) (*pkg.SearchQuery, error) {
	query := elastic.NewBoolQuery()
	if gender, ok := q["gender"]; ok {
		query = query.Filter(pkg.NewTermsString("gender", strings.Split(gender, ",")))
	}
	if title, ok := q["title"]; ok {
		query = query.Filter(pkg.NewTermsString("title", strings.Split(title, ",")))
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
		return &pkg.SearchQuery{
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
