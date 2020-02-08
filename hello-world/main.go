package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")

	// ElasticsearchAddress host name
	ElasticsearchAddress = os.Getenv("ELASTICSEARCH_SERVICE_HOST_NAME")
)

// V4Signer is a http.RoundTripper implementation to sign requests according to
// https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html. Many libraries allow customizing the behavior
// of HTTP requests, providing a transport. A V4Signer transport can be instantiated as follow:
//
// 	cfg, err := external.LoadDefaultAWSConfig()
//	if err != nil {
//		...
//	}
//	transport := &V4Signer{
//		RoundTripper: http.DefaultTransport,
//		Credentials:  cfg.Credentials,
//		Region:       cfg.Region,
//	}
type V4Signer struct {
	RoundTripper http.RoundTripper
	Credentials  aws.CredentialsProvider
	Region       string
}

// RoundTrip function
func (s *V4Signer) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	signer := v4.NewSigner(s.Credentials)
	switch req.Body {
	case nil:
		_, err := signer.Sign(ctx, req, nil, "es", s.Region, time.Now())
		if err != nil {
			return nil, fmt.Errorf("error signing request: %w", err)
		}
	default:
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_, err = signer.Sign(ctx, req, bytes.NewReader(b), "es", s.Region, time.Now())
		if err != nil {
			return nil, fmt.Errorf("error signing request: %w", err)
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(b))
	}
	return s.RoundTripper.RoundTrip(req)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: 500,
		}, err
	}
	es, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: []string{
			ElasticsearchAddress,
		},
		Transport: &V4Signer{
			RoundTripper: http.DefaultTransport,
			Credentials:  cfg.Credentials,
			Region:       cfg.Region,
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: 500,
		}, err
	}
	fmt.Println(es.Info())

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": map[string]interface{}{
					"terms": []map[string]string{
						{"item_id": "10000"},
					},
				},
			},
		},
		"from": 0,
		"size": 36,
		"sort": map[string]interface{}{
			"updated_at": map[string]string{
				"order": "desc",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: 500,
		}, err
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("items"),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: 500,
		}, err
	}
	defer res.Body.Close()
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("error"),
				StatusCode: 500,
			}, err
		} else {
			return events.APIGatewayProxyResponse{
				Body: fmt.Sprintf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				),
				StatusCode: 500,
			}, nil
		}
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: 500,
		}, err
	}

	return events.APIGatewayProxyResponse{
		Body: fmt.Sprintf("[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
