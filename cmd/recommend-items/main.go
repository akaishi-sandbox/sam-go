package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	recommenditems "github.com/akaishi-sandbox/sam-go/internal/recommend-items"
	"github.com/akaishi-sandbox/sam-go/pkg"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
	es, err := pkg.NewElasticsearch(ctx, ElasticsearchAddress)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	fmt.Println(es.Info())

	buf, err := recommenditems.CreateSourceItem(request.QueryStringParameters)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	fmt.Printf("query search: %s\n", buf.String())

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
	r, err := pkg.ElasticsearchParse(res)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	buf, err = recommenditems.CreateRecommendItems(r["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{}), request.QueryStringParameters)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	res, err = es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex("items"),
		es.Search.WithBody(&buf),
	)

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
