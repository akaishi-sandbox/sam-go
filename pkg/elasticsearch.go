package pkg

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	aws "github.com/olivere/elastic/aws/v4"
	elastic "github.com/olivere/elastic/v7"
)

// SearchQuery type
type SearchQuery struct {
	Index    string
	Query    *elastic.BoolQuery
	SortInfo elastic.SortInfo
	From     int
	Size     int
}

// Search function
func (sq *SearchQuery) Search(ctx context.Context, es *elastic.Client) (*elastic.SearchResult, error) {
	if len(sq.SortInfo.Field) == 0 {
		return es.Search().
			Index(sq.Index).
			Query(sq.Query).
			From(sq.From).
			Size(sq.Size). // take documents from-(size-from)
			Pretty(true).  // pretty print request and response JSON
			Do(ctx)
	}
	return es.Search().
		Index(sq.Index).
		Query(sq.Query).
		SortWithInfo(sq.SortInfo).
		From(sq.From).
		Size(sq.Size). // take documents from-(size-from)
		Pretty(true).  // pretty print request and response JSON
		Do(ctx)
}

// NewTermsString []string
func NewTermsString(name string, input []string) *elastic.TermsQuery {
	values := make([]interface{}, len(input))
	for i, s := range input {
		values[i] = s
	}
	return elastic.NewTermsQuery(name, values...)
}

// NewElasticsearch instance
func NewElasticsearch(ctx context.Context, elasticsearchAddress string) (*elastic.Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	signingClient := aws.NewV4SigningClient(sess.Config.Credentials, *sess.Config.Region)
	// From our experience, you should simply disable sniffing and health checks when using AWS Elasticsearch Service as it will do load-balancing on the server-side. Here's an example code of how this could be done
	return elastic.NewClient(
		elastic.SetURL(elasticsearchAddress),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(signingClient),
	)
}
