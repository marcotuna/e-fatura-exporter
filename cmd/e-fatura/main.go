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

	fmt.Printf("Launching e-fatura-exporter\nVersion %s\n\n", "1.0.0")

	// Get CSRF Token
	fmt.Println("Retriving CSRF Token")

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
	csrfDocument, err := goquery.NewDocumentFromReader(bytes.NewReader(httpClientGetCSRFToken.Body))
	if err != nil {
		//log.Fatal("Error loading HTTP response body. ", err)
	}

	csrfToken, ok := csrfDocument.Find(`input[name="_csrf"]`).First().Attr("value")
	if !ok {
		fmt.Println("Could not get CSRF")
		return
	}

	// Authenticate

	fmt.Println("Perform Authentication")

	httpClientAuth, err := service.HTTPClientReq(
		"https://www.acesso.gov.pt/jsp/submissaoFormularioLogin",
		url.Values{
			"path":               []string{"painelAdquirente.action"},
			"partID":             []string{"EFPF"},
			"authVersion":        []string{"1"},
			"_csrf":              []string{csrfToken},
			"selectedAuthMethod": []string{"N"},
			"username":           []string{os.Getenv("USERNAME")},
			"password":           []string{os.Getenv("PASSWORD")},
		},
		[]*service.HTTPClientHeader{
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
		},
		httpClientGetCSRFToken.Cookie,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	if httpClientAuth.StatusCode < 200 || httpClientAuth.StatusCode > 399 {
		fmt.Println("Authentication Failed.")
		return
	}

	// Create a goquery document from the HTTP response
	authDocument, err := goquery.NewDocumentFromReader(bytes.NewReader(httpClientAuth.Body))
	if err != nil {
		//log.Fatal("Error loading HTTP response body. ", err)
	}

	authErrMessage := authDocument.Find(`div[class="error-message"]`).Text()

	if authErrMessage != "" {
		fmt.Println(authErrMessage)
		return
	}

	// Extract user details from HTML Body
	authSign, _ := authDocument.Find(`input[name="sign"]`).First().Attr("value")
	authUserID, _ := authDocument.Find(`input[name="userID"]`).First().Attr("value")
	authSessionID, _ := authDocument.Find(`input[name="sessionID"]`).First().Attr("value")
	authNif, _ := authDocument.Find(`input[name="nif"]`).First().Attr("value")
	authTc, _ := authDocument.Find(`input[name="tc"]`).First().Attr("value")
	authTv, _ := authDocument.Find(`input[name="tv"]`).First().Attr("value")
	authUserName, _ := authDocument.Find(`input[name="userName"]`).First().Attr("value")
	authPartID, _ := authDocument.Find(`input[name="partID"]`).First().Attr("value")

	//fmt.Printf("Authentication Details: %v, %v, %v, %v, %v, %v, %v, %v", authSign, authUserID, authSessionID, authNif, authTc, authTv, authUserName, authPartID)

	// Get with Cookie
	fmt.Println("Retrive Invoices")

	httpClient, err := service.HTTPClientReq(
		"https://faturas.portaldasfinancas.gov.pt/consultarDocumentosAdquirente.action",
		url.Values{
			"sign":      []string{authSign},
			"userID":    []string{authUserID},
			"sessionID": []string{authSessionID},
			"nif":       []string{authNif},
			"tc":        []string{authTc},
			"tv":        []string{authTv},
			"userName":  []string{authUserName},
			"partID":    []string{authPartID},
		},
		[]*service.HTTPClientHeader{
			&service.HTTPClientHeader{
				Key:   "Content-Type",
				Value: []string{"application/x-www-form-urlencoded"},
			},
			&service.HTTPClientHeader{
				Key:   "Host",
				Value: []string{"faturas.portaldasfinancas.gov.pt"},
			},
			&service.HTTPClientHeader{
				Key:   "User-Agent",
				Value: []string{userAgent},
			},
			&service.HTTPClientHeader{
				Key:   "Upgrade-Insecure-Requests",
				Value: []string{"1"},
			},
		},
		httpClientAuth.Cookie,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v", string(httpClient.Body))
}
