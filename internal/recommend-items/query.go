package classificationinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CreateRecommendItems elastic search query
func CreateRecommendItems(r map[string]interface{}, q map[string]string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var filter []map[string]interface{}
	var must_not []map[string]interface{}
	if itemID, ok := q["item_id"]; ok {
		must_not = append(must_not, map[string]interface{}{
			"terms": map[string][]string{
				"item_id": strings.Split(itemID, ","),
			},
		})
	}
	if brand, ok := q["brand"]; ok {
		filter = append(must_not, map[string]interface{}{
			"terms": map[string][]string{
				"brand": strings.Split(brand, ","),
			},
		})
	}
	if gender, ok := r["gender"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"gender": strings.Split(gender.(string), ","),
			},
		})
	}
	if category, ok := r["category"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"category": strings.Split(category.(string), ","),
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

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter":   filter,
				"must_not": must_not,
			},
		},
		"from": from,
		"size": size,
		"sort": map[string]map[string]string{
			"updated_at": map[string]string{
				"order": "desc",
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return buf, err
	}

	return buf, nil
}

// CreateSourceItem elastic search query
func CreateSourceItem(q map[string]string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var filter []map[string]interface{}
	itemID, ok := q["item_id"]
	if !ok {
		return buf, fmt.Errorf("parameter not found")
	}
	filter = append(filter, map[string]interface{}{
		"terms": map[string][]string{
			"item_id": strings.Split(itemID, ","),
		},
	})

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
			},
		},
		"from": 0,
		"size": 1,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return buf, err
	}

	return buf, nil
}
