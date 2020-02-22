package database

import (
	"encoding/json"
	"testing"

	"github.com/akaishi-sandbox/sam-go/domain"
)

func TestCreateSearchQuery(t *testing.T) {
	testCase := func(q map[string]string, ok string) {
		query := createSearchQuery(q)

		if query.Index != "items" {
			t.Errorf("index error:%s", query.Index)
		}
		if query.Query == nil {
			t.Errorf("query nil")
		}
		s, err := query.Query.Source()
		if err != nil {
			t.Errorf("query source:%v", err)
		}
		j, err := json.Marshal(s)
		if err != nil {
			t.Errorf("query source not map string:%v", s)
		}
		source := string(j)

		if source != ok {
			t.Errorf("query source not map string:%s <> %s", source, ok)
		}
	}

	testCase(map[string]string{
		"aaa": "bbb",
	}, `{"bool":{}}`)
	testCase(map[string]string{
		"item_id": "123456AA",
	}, `{"bool":{"filter":{"terms":{"item_id":["123456AA"]}}}}`)
	testCase(map[string]string{
		"item_id": "123456AA",
		"gender":  "MEN",
	}, `{"bool":{"filter":[{"terms":{"item_id":["123456AA"]}},{"terms":{"gender":["MEN"]}}]}}`)
	testCase(map[string]string{
		"item_id":         "123456AA",
		"gender":          "MEN",
		"brand":           "UNIQLO",
		"category":        "シャツ",
		"discount_flag":   "1",
		"min_price":       "100",
		"max_price":       "10000",
		"exclude_expired": "1",
	}, `{"bool":{"filter":[{"terms":{"item_id":["123456AA"]}},{"terms":{"gender":["MEN"]}},{"terms":{"brand":["UNIQLO"]}},{"terms":{"category":["シャツ"]}},{"terms":{"discount_flag":["1"]}},{"range":{"lowest_price":{"from":100,"include_lower":true,"include_upper":true,"to":null}}},{"range":{"lowest_price":{"from":null,"include_lower":true,"include_upper":true,"to":10000}}}]}}`)
	testCase(map[string]string{
		"keywords": "UNIQLO",
	}, `{"bool":{"must":{"match_phrase":{"search_text":{"query":"UNIQLO"}}}}}`)
	testCase(map[string]string{
		"keywords": "UNIQLO シャツ",
	}, `{"bool":{"must":[{"match_phrase":{"search_text":{"query":"UNIQLO"}}},{"match_phrase":{"search_text":{"query":"シャツ"}}}]}}`)
}

func TestCreateRecommendItems(t *testing.T) {
	testCase := func(item domain.Item, q map[string]string, ok string) {
		query := createRecommendItems(item, q)

		if query.Index != "items" {
			t.Errorf("index error:%s", query.Index)
		}
		if query.Query == nil {
			t.Errorf("query nil")
		}
		s, err := query.Query.Source()
		if err != nil {
			t.Errorf("query source:%v", err)
		}
		j, err := json.Marshal(s)
		if err != nil {
			t.Errorf("query source not map string:%v", s)
		}
		source := string(j)

		if source != ok {
			t.Errorf("query source not map string:%s <> %s", source, ok)
		}
	}

	testCase(domain.Item{
		ItemID:   "ABCDEF",
		Gender:   "MEN",
		Category: "シャツ",
	}, map[string]string{
		"item_id": "ABCDEF",
	}, `{"bool":{"filter":[{"terms":{"gender":["MEN"]}},{"terms":{"category":["シャツ"]}}],"must_not":{"term":{"item_id":"ABCDEF"}}}}`)
}

func TestCreateClassificationQuery(t *testing.T) {
	testCase := func(q map[string]string, index, ok string) {
		query, err := createClassificationQuery(q)
		if err != nil {
			t.Errorf("createClassificationQuery error:%v", err)
		}

		if query.Index != index {
			t.Errorf("index error:%s <> %s", query.Index, index)
		}
		if query.Query == nil {
			t.Errorf("query nil")
		}
		s, err := query.Query.Source()
		if err != nil {
			t.Errorf("query source:%v", err)
		}
		j, err := json.Marshal(s)
		if err != nil {
			t.Errorf("query source not map string:%v", s)
		}
		source := string(j)

		if source != ok {
			t.Errorf("query source not map string:%s <> %s", source, ok)
		}
	}

	testCase(map[string]string{
		"index": "categories",
	},
		"categories",
		`{"bool":{}}`)
	testCase(map[string]string{
		"index": "brands",
		"title": "UNIQLO",
	},
		"brands",
		`{"bool":{"filter":{"terms":{"title":["UNIQLO"]}}}}`)
}
