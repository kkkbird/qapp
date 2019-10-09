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

// Get method
func Get(url string) (resp *http.Response, err error) {
	return QHTTPClient.Get(url)
}

// Post method
func Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return QHTTPClient.Post(url, contentType, body)
}

// PostForm method
func PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return QHTTPClient.PostForm(url, data)
}

// Head method
func Head(url string) (resp *http.Response, err error) {
	return QHTTPClient.Head(url)
}

// PostJSON method
func PostJSON(url string, body interface{}, result interface{}) (resp *http.Response, err error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	rsp, err := QHTTPClient.Post(url, "application/json", bytes.NewReader(b))
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
