package main

import (
	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"io"
	"strconv"
)

var (
	reviews     []string
	pageReviews []string
)

func FetchReviewsPerPage(url string) ([]string, error) {
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

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header = http.Header{
		"user-agent": {"Mozilla/5.0 ... Chrome/128.0.0.0 Safari/537.36"},
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	// HTML parser
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract reviews
	doc.Find(".a-section.review.aok-relative").Each(func(i int, s *goquery.Selection) {
		review := s.Find("div.a-row.a-spacing-small.review-data > span > span").Text()
		if review != "" {
			reviews = append(reviews, review)
		}
	})
	return reviews, nil
}
func FetchProductReviews(asin string, reviewCount int) ([]string, error) {
	url := "https://www.amazon.de/-/en/product-reviews/" + asin + "??ie=UTF8&reviewerType=all_reviews"
	pageReviews, err = FetchReviewsPerPage(url)
	if err != nil {
		return nil, err
	}

	var pageNumber = 2
	for len(pageReviews) < reviewCount {
		url = url + "&pageNumber=" + strconv.Itoa(pageNumber)
		extraPageReviews, err := FetchReviewsPerPage(url)
		if err != nil {
			return nil, err
		}
		pageReviews = append(pageReviews, extraPageReviews...)
		if len(pageReviews) < reviewCount {
			pageNumber++
			continue
		} else {
			break
		}
	}
	return pageReviews[:reviewCount], nil
}
