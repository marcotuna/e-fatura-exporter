package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/marcotuna/e-fatura-exporter/service"
)

const (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36"
)

func main() {

	// Get CSRF Token
	httpClientGetCSRFToken, err := service.HTTPClientReq(
		"https://www.acesso.gov.pt/jsp/loginRedirectForm.jsp?path=painelAdquirente.action&partID=EFPF",
		url.Values{},
		[]*service.HTTPClientHeader{
			&service.HTTPClientHeader{
				Key:   "User-Agent",
				Value: []string{userAgent},
			},
		},
		nil,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(httpClientGetCSRFToken.Body))
	if err != nil {
		//log.Fatal("Error loading HTTP response body. ", err)
	}

	csrfToken, ok := document.Find(`input[name="_csrf"]`).First().Attr("value")
	if !ok {
		fmt.Println("Could not get CSRF")
		return
	}

	// Authenticate
	httpClientHeaders := []*service.HTTPClientHeader{
		&service.HTTPClientHeader{
			Key:   "Content-Type",
			Value: []string{"application/x-www-form-urlencoded"},
		},
		&service.HTTPClientHeader{
			Key:   "Host",
			Value: []string{"www.acesso.gov.pt"},
		},
		&service.HTTPClientHeader{
			Key:   "User-Agent",
			Value: []string{userAgent},
		},
		&service.HTTPClientHeader{
			Key:   "Upgrade-Insecure-Requests",
			Value: []string{"1"},
		},
	}

	httpClientPostParams := url.Values{
		"path":               []string{"painelAdquirente.action"},
		"partID":             []string{"EFPF"},
		"authVersion":        []string{"1"},
		"_csrf":              []string{csrfToken},
		"selectedAuthMethod": []string{"N"},
		"username":           []string{os.Getenv("USERNAME")},
		"password":           []string{os.Getenv("PASSWORD")},
	}

	httpClientAuthResp, err := service.HTTPClientReq("https://www.acesso.gov.pt/jsp/submissaoFormularioLogin", httpClientPostParams, httpClientHeaders, httpClientGetCSRFToken.Cookie)
	if err != nil {
		fmt.Println(err)
		return
	}

	httpClient, err := service.HTTPClientReq(
		"https://faturas.portaldasfinancas.gov.pt/consultarDocumentosAdquirente.action",
		url.Values{},
		httpClientAuthResp.Header,
		httpClientAuthResp.Cookie,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v", httpClient.Cookie)
	fmt.Printf("%+v", string(httpClient.Body))
	//fmt.Printf("%+v", httpClientAuthResp.Header)
}
