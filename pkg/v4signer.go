package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
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
//		Context:      ctx,
//	}
type V4Signer struct {
	RoundTripper http.RoundTripper
	Credentials  aws.CredentialsProvider
	Region       string
	Context      context.Context
}

// RoundTrip function
func (s *V4Signer) RoundTrip(req *http.Request) (*http.Response, error) {
	signer := v4.NewSigner(s.Credentials)
	switch req.Body {
	case nil:
		_, err := signer.Sign(s.Context, req, nil, "es", s.Region, time.Now())
		if err != nil {
			return nil, fmt.Errorf("error signing request: %w", err)
		}
	default:
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_, err = signer.Sign(s.Context, req, bytes.NewReader(b), "es", s.Region, time.Now())
		if err != nil {
			return nil, fmt.Errorf("error signing request: %w", err)
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(b))
	}
	return s.RoundTripper.RoundTrip(req)
}
