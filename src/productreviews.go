package main

import (
	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func FetchProductReviews(url string) ([]string, error) {
	var reviews []string

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
		return nil, err
	}

	// HTTP-Anfrage erstellen
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Header hinzufügen
	req.Header = http.Header{
		"user-agent": {"Mozilla/5.0 ... Chrome/128.0.0.0 Safari/537.36"},
	}

	// Anfrage ausführen
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// HTML parsen
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Bewertungen extrahieren
	doc.Find(".a-section.review.aok-relative").Each(func(i int, s *goquery.Selection) {
		review := s.Find("div.a-row.a-spacing-small.review-data > span > span").Text()
		if review != "" {
			reviews = append(reviews, review)
		}
	})

	if len(reviews) == 0 {
		return nil, nil
	}

	return reviews, nil
}
