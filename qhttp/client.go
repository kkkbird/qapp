package qhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
)

// WithAuthorization add authorization bearer to request
func WithAuthorization(token string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// Get method
func Get(url string, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for _, f := range reqOpts {
		f(req)
	}

	return QHTTPClient.Do(req)
}

// Post method
func Post(url, contentType string, body io.Reader, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	for _, f := range reqOpts {
		f(req)
	}

	return QHTTPClient.Do(req)
}

// PostForm method
func PostForm(url string, data url.Values, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	return Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), reqOpts...)
}

// Head method
func Head(url string, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	for _, f := range reqOpts {
		f(req)
	}
	return QHTTPClient.Do(req)
}

// PostJSON method
func PostJSON(url string, body interface{}, result interface{}, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	rsp, err := Post(url, "application/json", bytes.NewReader(b), reqOpts...)

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
