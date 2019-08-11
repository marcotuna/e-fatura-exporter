package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/marcotuna/e-fatura-exporter/utils"
	"golang.org/x/net/publicsuffix"
)

// HTTPClientHeader Headers for HTTP Client
type HTTPClientHeader struct {
	Key   string
	Value []string
}

// HTTPClientBody Body for HTTP Client
type HTTPClientBody struct {
	Key   string
	Value string
}

// HTTPClientResponse ...
type HTTPClientResponse struct {
	Body       []byte
	Cookie     []*http.Cookie
	Header     []*HTTPClientHeader
	StatusCode int
}

// HTTPClientReq ...
func HTTPClientReq(clientURL string, postParams url.Values, reqHeaders []*HTTPClientHeader, reqCookies []*http.Cookie) (*HTTPClientResponse, error) {

	var req *http.Request
	var err error

	if len(postParams) > 0 {
		req, err = http.NewRequest("POST", clientURL, bytes.NewBufferString(postParams.Encode()))
	} else {
		req, err = http.NewRequest("GET", clientURL, nil)
	}

	// Pass received headers
	if len(reqHeaders) > 0 {
		for _, v := range reqHeaders {
			req.Header.Set(v.Key, v.Value[0])
		}
	}

	// Content-Type JSON
	//req.Header.Set("Content-Type", "application/json")

	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return &HTTPClientResponse{}, err
	}

	// Pass received cookies
	cookieJar.SetCookies(req.URL, reqCookies)

	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPClientResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Set Cookies
	var httpCookie []*http.Cookie

	maxAge := 7200
	expiresAt := time.Unix(utils.GetMillis()/1000+int64(maxAge), 0)

	secure := false
	if utils.GetProtocol(req) == "https" {
		secure = true
	}

	for _, v := range cookieJar.Cookies(req.URL) {
		httpCookie = append(httpCookie, &http.Cookie{
			Name:     v.Name,
			Value:    v.Value,
			Path:     v.Path,
			Domain:   v.Domain,
			Secure:   secure,
			HttpOnly: v.HttpOnly,
			MaxAge:   maxAge,
			Expires:  expiresAt,
		})
	}

	// Set Headers
	httpHeaders := []*HTTPClientHeader{}

	if resp.Header != nil {
		for k, v := range resp.Header {
			switch k {
			case "Set-Cookie":

			default:
				httpHeaders = append(httpHeaders, &HTTPClientHeader{Key: k, Value: v})
			}

		}
	}

	return &HTTPClientResponse{Body: body, Cookie: httpCookie, Header: httpHeaders, StatusCode: resp.StatusCode}, nil
}
