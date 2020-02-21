package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	searchitems "github.com/akaishi-sandbox/sam-go/internal/search-items"
	"github.com/akaishi-sandbox/sam-go/pkg"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	elastic "github.com/olivere/elastic/v7"
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

	query, err := searchitems.CreateSearchItems(request.QueryStringParameters)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	searchResult, err := query.Search(ctx, es)
	if err != nil {
		sentry.CaptureException(err)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error"),
			StatusCode: http.StatusServiceUnavailable,
		}, err
	}

	// 更新元の商品はIDを元に検索しているので複数個存在する場合がある、そのためアクセス回数の最も大きい値を更新元の数字として取得する
	numberOfAccess := 0
	var iType searchitems.Item
	for _, item := range searchResult.Each(reflect.TypeOf(iType)) {
		if i, ok := item.(searchitems.Item); ok {
			if i.AccessCounter > numberOfAccess {
				numberOfAccess = i.AccessCounter
			}
		}
	}
	numberOfAccess++
	lastAccessedAt := time.Now()

	for _, hit := range searchResult.Hits.Hits {
		es.Update().Index(hit.Index).Id(hit.Id).
			Script(elastic.NewScript("ctx._source.access_counter = params.access_counter").Param("access_counter", numberOfAccess)).
			Script(elastic.NewScript("ctx._source.last_accessed_at = params.last_accessed_at").Param("last_accessed_at", lastAccessedAt)).
			Do(ctx)
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(struct {
		Total int64                `json:"total"`
		Hits  []*elastic.SearchHit `json:"hits"`
	}{
		Total: searchResult.TotalHits(),
		Hits:  searchResult.Hits.Hits,
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
	defer sentry.Flush(time.Second * 5)
	defer sentry.Recover()
	lambda.Start(handler)
}
