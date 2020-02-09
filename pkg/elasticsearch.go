package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// NewElasticsearch instance
func NewElasticsearch(ctx context.Context, elasticsearchAddress string) (*elasticsearch7.Client, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	return elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: []string{
			elasticsearchAddress,
		},
		Transport: &V4Signer{
			RoundTripper: http.DefaultTransport,
			Credentials:  cfg.Credentials,
			Region:       cfg.Region,
			Context:      ctx,
		},
	})
}

// ElasticsearchParse error parse
func ElasticsearchParse(res *esapi.Response) (map[string]interface{}, error) {
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r, nil
}
