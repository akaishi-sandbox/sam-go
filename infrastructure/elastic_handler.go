package infrastructure

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	aws "github.com/olivere/elastic/aws/v4"
	elastic "github.com/olivere/elastic/v7"
)

type ElasticHandler struct {
	Client  *elastic.Client
	Context context.Context
}

type ElasticQuery struct {
	Index    string
	Query    *elastic.BoolQuery
	SortInfo elastic.SortInfo
	From     int
	Size     int
}

func (handler *ElasticHandler) Search(eq *ElasticQuery) (*elastic.SearchResult, error) {
	return handler.Client.Search().
		Index(eq.Index).
		Query(eq.Query).
		SortWithInfo(eq.SortInfo).
		From(eq.From).
		Size(eq.Size). // take documents from-(size-from)
		Pretty(true).  // pretty print request and response JSON
		Do(handler.Context)
}

// NewElasticHandler instance
func NewElasticHandler(ctx context.Context, elasticsearchAddress string) (*ElasticHandler, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	signingClient := aws.NewV4SigningClient(sess.Config.Credentials, *sess.Config.Region)
	// From our experience, you should simply disable sniffing and health checks when using AWS Elasticsearch Service as it will do load-balancing on the server-side. Here's an example code of how this could be done
	es, err := elastic.NewClient(
		elastic.SetURL(elasticsearchAddress),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(signingClient),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
	)
	if err != nil {
		return nil, err
	}

	return &ElasticHandler{
		Client:  es,
		Context: ctx,
	}, nil
}
