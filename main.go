package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.amazon.de/Devoko-H%C3%B6henverstellbarer-Computertisch-2-Funktions-Memory-Ergonomischer/dp/B0CXY2M22T/?th=1", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"sec-ch-ua":                 {"\"Chromium\";v=\"128\", \"Not;A=Brand\";v=\"24\", \"Google Chrome\";v=\"128\""},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"\"Windows\""},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br, zstd"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"priority":                  {"u=0, i"},
		http.HeaderOrderKey:         {"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language", "priority"},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("status code: %d\n", resp.StatusCode)

	// HTML parsen
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Titel extrahieren
	title := strings.TrimSpace(doc.Find("#productTitle").Text())
	if title == "" {
		title = "Titel nicht gefunden"
	}

	// Beschreibung extrahieren
	description := strings.TrimSpace(doc.Find("#productDescription").Text())
	if description == "" {
		// Alternative Beschreibung finden (Feature-Bullets)
		description = strings.TrimSpace(doc.Find("#feature-bullets").Text())
		if description == "" {
			description = "Beschreibung nicht gefunden"
		}
	}

	// Bewertungen extrahieren
	reviews := strings.TrimSpace(doc.Find("#acrCustomerReviewText").Text())
	if reviews == "" {
		reviews = "Bewertungen nicht gefunden"
	}

	fmt.Println("Title: " + title)

	fmt.Println("Description: " + description)

	fmt.Println("Reviews: " + reviews)

	req, err = http.NewRequest(http.MethodGet, "https://www.amazon.de/Devoko-H%C3%B6henverstellbarer-Computertisch-2-Funktions-Memory-Ergonomischer/product-reviews/B0CXY2M22T/ref=cm_cr_getr_d_paging_btm_prev_1?ie=UTF8&reviewerType=all_reviews", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"sec-ch-ua":                 {"\"Chromium\";v=\"128\", \"Not;A=Brand\";v=\"24\", \"Google Chrome\";v=\"128\""},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"\"Windows\""},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br, zstd"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"priority":                  {"u=0, i"},
		http.HeaderOrderKey:         {"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language", "priority"},
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("status code: %d\n", resp.StatusCode)

	// HTML parsen
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(".a-section.review.aok-relative").Each(func(i int, s *goquery.Selection) {
		review := s.Find("div.a-row.a-spacing-small.review-data > span > span")
		fmt.Println(review.Text())
	})
}
