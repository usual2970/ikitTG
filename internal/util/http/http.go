package http

import (
	"io"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
)

type Options struct {
	Timeout time.Duration
}

type Option func(o *Options)

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

func Req(url string, method string, body io.Reader, head map[string]string, opts ...Option) ([]byte, error) {
	options := &Options{
		Timeout: 30000 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(options)
	}
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(options.Timeout))

	// Create an http.Request instance
	req, _ := http.NewRequest(method, url, body)
	for k, v := range head {
		req.Header.Set(k, v)
	}
	// Call the `Do` method, which has a similar interface to the `http.Do` method
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
