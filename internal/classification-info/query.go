package classificationinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CreateSearchQuery elastic search query
func CreateSearchQuery(q map[string]string) (string, bytes.Buffer, error) {
	var filter []map[string]interface{}
	if gender, ok := q["gender"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"gender": strings.Split(gender, ","),
			},
		})
	}
	if title, ok := q["title"]; ok {
		filter = append(filter, map[string]interface{}{
			"terms": map[string][]string{
				"title": strings.Split(title, ","),
			},
		})
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
	sort := map[string]interface{}{
		"sort_no": map[string]string{
			"order": "asc",
		},
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

	var index string
	if index, ok := q["index"]; !ok {
		return index, buf, fmt.Errorf("parameter not found")
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return index, buf, err
	}

	return index, buf, nil
}
