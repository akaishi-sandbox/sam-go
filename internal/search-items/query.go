package searchitems

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// CreateSearchQuery elastic search query
func CreateSearchQuery(q map[string]string) (bytes.Buffer, error) {
	var filter []map[string]interface{}
	if itemID, ok := q["item_id"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"item_id": strings.Split(itemID, ","),
			},
		})
	}
	if gender, ok := q["gender"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"gender": strings.Split(gender, ","),
			},
		})
	}
	if brand, ok := q["brand"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"brand": strings.Split(brand, ","),
			},
		})
	}
	if category, ok := q["category"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"category": strings.Split(category, ","),
			},
		})
	}
	if discountFlag, ok := q["discount_flag"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"discount_flag": strings.Split(discountFlag, ","),
			},
		})
	}
	if minPrice, ok := q["min_price"]; ok {
		if price, err := strconv.Atoi(minPrice); err == nil {
			filter = append(filter, map[string]interface{}{
				"range": map[string]map[string]int{
					"lowest_price": map[string]int{
						"gte": price,
					},
				},
			})
		}
	}
	if maxPrice, ok := q["max_price"]; ok {
		if price, err := strconv.Atoi(maxPrice); err == nil {
			filter = append(filter, map[string]interface{}{
				"range": map[string]map[string]int{
					"lowest_price": map[string]int{
						"lte": price,
					},
				},
			})
		}
	}
	if keywords, ok := q["keywords"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"keywords": strings.Split(keywords, ","),
			},
		})
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
	sort := map[string]interface{}{
		"updated_at": map[string]string{
			"order": "desc",
		},
	}
	if order, ok := q["order"]; ok {
		switch order {
		case "new":
			sort = map[string]interface{}{
				"updated_at": map[string]string{
					"order": "desc",
				},
			}
		case "min-max":
			sort = map[string]interface{}{
				"lowest_price": map[string]string{
					"order": "asc",
				},
			}
		case "max-max":
			sort = map[string]interface{}{
				"lowest_price": map[string]string{
					"order": "desc",
				},
			}
		}
	}
	if excludeExpired, ok := q["exclude_expired"]; !ok && excludeExpired == "1" {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]int{
				"release_flag": {0, 1},
			},
		})
	}

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
			},
		},
		"from": from,
		"size": size,
		"sort": sort,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return buf, err
	}

	return buf, nil
}
