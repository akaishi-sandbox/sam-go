package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/akaishi-sandbox/sam-go/pkg"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/getsentry/sentry-go"
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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	es, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: []string{
			ElasticsearchAddress,
		},
		Transport: &pkg.V4Signer{
			RoundTripper: http.DefaultTransport,
			Credentials:  cfg.Credentials,
			Region:       cfg.Region,
			Context:      ctx,
		},
	})
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusNotImplemented,
		}, err
	}
	fmt.Println(es.Info())

	q := request.QueryStringParameters
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
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusBadGateway,
		}, err
	}

	fmt.Printf("query search:%s\n", buf.String())

	res, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex("items"),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusServiceUnavailable,
		}, err
	}
	defer res.Body.Close()
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			sentry.CaptureException(err)
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("error"),
				StatusCode: http.StatusGatewayTimeout,
			}, err
		} else {
			sentry.CaptureException(fmt.Errorf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"]))
			return events.APIGatewayProxyResponse{
				Body: fmt.Sprintf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				),
				StatusCode: http.StatusHTTPVersionNotSupported,
			}, nil
		}
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusVariantAlsoNegotiates,
		}, err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(struct {
		Total int `json:"total"`
		Items interface{}
	}{
		Total: int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		Items: r["hits"].(map[string]interface{})["hits"],
	}); err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusBadGateway,
		}, err
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type": "application/json;charset=UTF-8",
		},
		Body:       body.String(),
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	lambda.Start(handler)
}
