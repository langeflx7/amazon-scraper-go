package main

import (
	"github.com/langeflx7/amazonscrapergo/src/Model"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var (
	fetched_Information Model.Product
)

func FetchProductInfo(asin string) (Model.Product, error) {
	// HTTP-Client erstellen

	// Concatenate URL directly with product ASIN
	failedScrape := Model.Product{
		Title:       "",
		Description: "",
		Rating:      "",
		Price:       "",
	}
	url := "https://www.amazon.de/-/en/dp/" + asin
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return failedScrape, err
	}

	// HTTP-Anfrage erstellen
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return failedScrape, err
	}

	// Header hinzufügen
	req.Header = http.Header{
		"user-agent": {"Mozilla/5.0 ... Chrome/128.0.0.0 Safari/537.36"},
	}
	// Anfrage ausführen
	resp, err := client.Do(req)
	if err != nil {
		return failedScrape, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	// HTML parsen
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return failedScrape, err
	}

	// Produktinformationen extrahieren
	title := strings.TrimSpace(doc.Find("#productTitle").Text())
	if title == "" {
		title = "Titel nicht gefunden"
	}

	description := strings.TrimSpace(strings.TrimSpace(doc.Find("#productDescription").Text()))
	if description == "" {
		description = strings.TrimSpace(doc.Find("#feature-bullets").Text())
		if description == "" {
			description = "Beschreibung nicht gefunden"
		}
	}

	ratingClass := strings.TrimSpace(doc.Find("span.a-size-base.a-color-base").Text())
	words := strings.Fields(ratingClass)
	var rating string
	if len(words) > 0 {
		rating = words[0]
	}
	if rating == "" {
		rating = "Bewertung nicht gefunden"
	}

	priceClass := strings.TrimSpace(doc.Find("span.aok-offscreen").Text())
	priceArray := strings.Fields(priceClass)
	var price string
	if len(priceArray) > 0 {
		price = priceArray[0]
	}
	if price == "" {
		price = "Preis nicht gefunden"
	}
	fetched_Information = Model.Product{
		Title:       title,
		Description: description,
		Rating:      rating,
		Price:       price,
	}
	return fetched_Information, nil
}
