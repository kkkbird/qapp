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

	"github.com/gin-gonic/gin/binding"
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
func Get(uri string, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	for _, f := range reqOpts {
		f(req)
	}

	return QHTTPClient.Do(req)
}

// GetJSON method
func GetJSON(uri string, result interface{}, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
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
func Post(uri, contentType string, body io.Reader, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", uri, body)
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
func PostForm(uri string, data url.Values, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	return Post(uri, binding.MIMEPOSTForm, strings.NewReader(data.Encode()), reqOpts...)
}

// Head method
func Head(uri string, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", uri, nil)
	if err != nil {
		return nil, err
	}
	for _, f := range reqOpts {
		f(req)
	}
	return QHTTPClient.Do(req)
}

// PostJSON method
func PostJSON(uri string, body interface{}, result interface{}, reqOpts ...func(*http.Request)) (resp *http.Response, err error) {
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
