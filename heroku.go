package hbuild

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// unverifiedSSLTransport is a transport that mirrors http.DefaultTransport
// and skips SSL verification.
var unverifiedSSLTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
}

type herokuClient struct {
	httpClient *http.Client
	token      string
	url        *url.URL
	version    string
	userAgent  string
}

type herokuRequest struct {
	method            string
	path              string
	body              interface{}
	additionalHeaders http.Header
}

func newHerokuClient(token string) herokuClient {
	herokuHost := "https://api.heroku.com"
	if hh := os.Getenv("HEROKU_API_URL"); hh != "" {
		herokuHost = hh
	}

	transport := http.DefaultTransport
	if strings.Contains(herokuHost, "herokudev") {
		transport = unverifiedSSLTransport
	}

	herokuUrl, _ := url.Parse(herokuHost)
	return herokuClient{
		httpClient: &http.Client{Transport: transport},
		token:      token,
		url:        herokuUrl,
		version:    "application/vnd.heroku+json; version=edge",
		userAgent:  "hbuild/1",
	}
}

func (hc herokuClient) request(hrequest herokuRequest, v interface{}) (err error) {
	var requestBody io.Reader

	url := hc.url.String() + hrequest.path

	if hrequest.body != nil {
		requestJson, err := json.Marshal(hrequest.body)
		if err != nil {
			return err
		}
		// Patch the really crappy way api.h.c handles json escaping in
		// non-production environments.
		// TODO: get api.h.c to behave correctly in non-prod environments
		requestJson = []byte(strings.Replace(string(requestJson), "\\u0026", "&", -1))
		requestBody = bytes.NewReader(requestJson)
	}

	request, err := http.NewRequest(hrequest.method, url, requestBody)
	if err != nil {
		return
	}

	request.SetBasicAuth("", hc.token)
	request.Header.Set("Accept", hc.version)
	request.Header.Set("User-Agent", hc.userAgent)
	if hrequest.body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for k, v := range hrequest.additionalHeaders {
		request.Header[k] = v
	}

	response, err := hc.httpClient.Do(request)
	if err != nil {
		return
	}

	if response.StatusCode/100 != 2 {
		var herr HerokuJsonError
		err = json.NewDecoder(response.Body).Decode(&herr)
		if err != nil {
			return err
		}
		return HerokuError{errors.New(herr.Message), herr.Id, herr.URL}
	}

	err = json.NewDecoder(response.Body).Decode(&v)
	return
}

type HerokuJsonError struct {
	Message string
	Id      string
	URL     string `json:"url"`
}

type HerokuError struct {
	error
	Id  string
	URL string
}
