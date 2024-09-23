package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func FetchProductInfo(url string) (string, string, string, error) {
	// HTTP-Client erstellen
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return "", "", "", err
	}

	// HTTP-Anfrage erstellen
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", "", err
	}

	// Header hinzufügen
	req.Header = http.Header{
		"user-agent": {"Mozilla/5.0 ... Chrome/128.0.0.0 Safari/537.36"},
	}

	// Anfrage ausführen
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	// HTML parsen
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	// Produktinformationen extrahieren
	title := strings.TrimSpace(doc.Find("#productTitle").Text())
	if title == "" {
		title = "Titel nicht gefunden"
	}

	description := strings.TrimSpace(doc.Find("#productDescription").Text())
	if description == "" {
		description = strings.TrimSpace(doc.Find("#feature-bullets").Text())
		if description == "" {
			description = "Beschreibung nicht gefunden"
		}
	}

	reviewsSummary := strings.TrimSpace(doc.Find("#acrCustomerReviewText").Text())
	if reviewsSummary == "" {
		reviewsSummary = "Bewertungen nicht gefunden"
	}

	return title, description, reviewsSummary, nil
}
