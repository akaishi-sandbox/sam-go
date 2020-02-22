package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/akaishi-sandbox/sam-go/infrastructure"
	"github.com/akaishi-sandbox/sam-go/interfaces/controllers"
	"github.com/akaishi-sandbox/sam-go/pkg/sentryecho"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echolamda "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	ElasticsearchAddress = os.Getenv("ELASTICSEARCH_SERVICE_HOST_NAME")
)

var echoLambda *echolamda.EchoLambda

// Handler is the main entry point for Lambda. Receives a proxy request and
// returns a proxy response
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if echoLambda == nil {
		elasticHandler, err := infrastructure.NewElasticHandler(ctx, ElasticsearchAddress)
		if err != nil {
			sentry.CaptureException(err)
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("error"),
				StatusCode: http.StatusInternalServerError,
			}, err
		}

		e := echo.New()
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(sentryecho.New(sentryecho.Options{}))

		itemController := controllers.NewItemController(elasticHandler)

		e.GET("/search-items", itemController.Search)
		e.GET("/recommend-items", itemController.Recommend)
		e.GET("/classification-info", itemController.Classification)
		e.GET("/access-info", itemController.Access)
		echoLambda = echolamda.New(e)
	}

	return echoLambda.ProxyWithContext(ctx, req)
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	lambda.Start(Handler)
}
