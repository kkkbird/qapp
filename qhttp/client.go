package qhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

var (
	// QHTTPTransport is the default transport for qhttpclient
	QHTTPTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// QHTTPClient is the default client
	QHTTPClient = &http.Client{
		Transport: QHTTPTransport,
	}

	ErrLimitExceed = errors.New("limit exceed")
)

const (
	LimitNoBlock     = -1
	LimitAlwaysBlock = 0
)

// WithAuthorization add authorization bearer to request
func WithAuthorization(token string) func(*http.Request) error {
	return func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func WithBasicAuthorization(username, password string) func(*http.Request) error {
	return func(req *http.Request) error {
		req.SetBasicAuth(username, password)
		return nil
	}
}

type Limit struct {
	Limiter *redis_rate.Limiter
	redis_rate.Limit
	Block time.Duration // ms
	Key   string
}

func NewLimit(rdb *redis.Client, key string, rate int, period time.Duration, burst int, block time.Duration) *Limit {
	return &Limit{
		Limiter: redis_rate.NewLimiter(rdb),
		Limit: redis_rate.Limit{
			Rate:   rate,
			Period: period,
			Burst:  burst,
		},
		Block: block,
		Key:   key,
	}
}

func WithLimit(l *Limit) func(*http.Request) error {
	return func(req *http.Request) error {
		var ctx context.Context

		if l.Block <= 0 {
			ctx = context.Background()
		} else {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), l.Block)
			defer cancel()
		}

		var key = l.Key

		if key == "" {
			key = req.Host + "/" + req.RequestURI
		}

		key = "qhttp:" + key

		for {
			rlt, err := l.Limiter.Allow(ctx, key, l.Limit)
			if err != nil {
				return err
			}

			if rlt.Allowed > 0 {
				return nil
			}

			if l.Block < 0 { // <0 means cancel the request
				return ErrLimitExceed
			}

			select {
			case <-time.After(rlt.RetryAfter):

			case <-ctx.Done():
				return context.DeadlineExceeded
			}
		}
	}
}

// Get method
func Get(uri string, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	for _, f := range reqOpts {
		err := f(req)
		if err != nil {
			return nil, err
		}
	}

	return QHTTPClient.Do(req)
}

// GetJSON method
func GetJSON(uri string, result interface{}, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	rsp, err := Get(uri, reqOpts...)

	if err != nil {
		return nil, err
	}

	if result != nil { //if result != nil, try Unmarshal the body
		defer rsp.Body.Close()
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}

		if rsp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Status(%d):%s", rsp.StatusCode, rspBody)
		}

		err = json.Unmarshal(rspBody, result)
		if err != nil {
			return rsp, err
		}
	}

	return rsp, err
}

// Post method
func Post(uri, contentType string, body io.Reader, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	for _, f := range reqOpts {
		err := f(req)
		if err != nil {
			return nil, err
		}
	}

	return QHTTPClient.Do(req)
}

// PostForm method
func PostForm(uri string, data url.Values, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	return Post(uri, binding.MIMEPOSTForm, strings.NewReader(data.Encode()), reqOpts...)
}

// Head method
func Head(uri string, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", uri, nil)
	if err != nil {
		return nil, err
	}
	for _, f := range reqOpts {
		err := f(req)
		if err != nil {
			return nil, err
		}
	}
	return QHTTPClient.Do(req)
}

// PostJSON method
func PostJSON(uri string, body interface{}, result interface{}, reqOpts ...func(*http.Request) error) (resp *http.Response, err error) {
	var (
		_b          []byte
		contentType string
	)

	if _body, ok := body.(url.Values); ok {
		_b = []byte(_body.Encode())
		contentType = binding.MIMEPOSTForm
	} else {
		_b, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		contentType = binding.MIMEJSON
	}

	rsp, err := Post(uri, contentType, bytes.NewReader(_b), reqOpts...)

	if err != nil {
		return nil, err
	}

	if result != nil { //if result != nil, try Unmarshal the body
		defer rsp.Body.Close()
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}

		if rsp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Status(%d):%s", rsp.StatusCode, rspBody)
		}

		err = json.Unmarshal(rspBody, result)
		if err != nil {
			return rsp, err
		}
	}

	return rsp, err
}
