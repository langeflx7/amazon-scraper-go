package productviewer

import (
	"io"
	"strings"

	"github.com/langeflx7/amazonscrapergo/src/Model"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var (
	fetchedInformation Model.Product
)

func FetchProductInfo(asin string) (Model.Product, error) {

	failedScrape := Model.Product{
		Title:       "",
		Description: "",
		Rating:      "",
		Price:       "",
	}
	// Concatenate URL directly with product ASIN
	url := "https://www.amazon.de/-/en/dp/" + asin
	jar := tls_client.NewCookieJar()

	// Create HTTP client
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

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return failedScrape, err
	}

	// Add headers
	req.Header = http.Header{
		"user-agent": {"Mozilla/5.0 ... Chrome/128.0.0.0 Safari/537.36"},
	}
	// Execute request
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

	// HTML parser
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return failedScrape, err
	}

	// Product Information extraction
	title := strings.TrimSpace(doc.Find("#productTitle").Text())
	if title == "" {
		title = "Title not found"
	}

	description := strings.TrimSpace(strings.TrimSpace(doc.Find("#productDescription").Text()))
	if description == "" {
		description = strings.TrimSpace(doc.Find("#feature-bullets").Text())
		if description == "" {
			description = "Description not found"
		}
	}

	ratingClass := strings.TrimSpace(doc.Find("span.a-size-base.a-color-base").Text())
	words := strings.Fields(ratingClass)
	var rating string
	if len(words) > 0 {
		rating = words[0]
	}
	if rating == "" {
		rating = "Rating not found"
	}

	var price string
	priceClass := strings.TrimSpace(doc.Find("span.aok-offscreen").Text())
	if priceClass != "" {
		if strings.Contains(priceClass, "&euro;") {
			cleanPriceClass := strings.ReplaceAll(priceClass, "&euro;", "")
			price = "€" + strings.Join(strings.Split(cleanPriceClass, " ")[:3], " ")
		} else {
			price = strings.Fields(priceClass)[0]
		}
	} else {
		priceFromButton := strings.TrimSpace(doc.Find("span.a-size-mini.olpWrapper").Text())
		price = priceFromButton
		if price == "" {
			doc.Find("span.a-price.a-text-price.a-size-medium").Each(func(i int, s *goquery.Selection) {
				price = "€" + strings.Split(s.Text(), "€")[1]
			})
		}
	}
	var extraInfoCategories []string
	doc.Find("ul.a-unordered-list.a-nostyle.a-vertical.a-spacing-none.detail-bullet-list").Each(func(i int, s *goquery.Selection) {
		extraProductInfo := strings.TrimSpace(s.Find("li").Text())
		if extraProductInfo != "" {
			extraInfoCategories = append(extraInfoCategories, extraProductInfo)
		}
	})
	var extraInfoString string
	if len(extraInfoCategories) != 0 {
		extraInfoString = strings.Join(strings.Fields(strings.Join(extraInfoCategories[:2], "")), " ")
	}

	var category, categoryValue string
	var categoryArray, categoryValueArray []string
	if extraInfoString == "" && doc.Find("#productDetails_techSpec_section_1").Text() != "" {
		doc.Find("th.a-color-secondary.a-size-base.prodDetSectionEntry").Each(func(i int, s *goquery.Selection) {
			category = strings.TrimSpace(s.Text())
			if category != "" && category != "Customer Reviews" && category != "Best Sellers Rank" {
				categoryArray = append(categoryArray, category)
			}
		})
		doc.Find("td.a-size-base.prodDetAttrValue").Each(func(i int, s *goquery.Selection) {
			categoryValue = strings.TrimSpace(s.Text())
			if categoryValue != "" {
				categoryValueArray = append(categoryValueArray, categoryValue)
			}
		})
	}
	if len(categoryArray) != len(categoryValueArray) {
		categoryValueArray = categoryValueArray[:len(categoryArray)]
	}
	for i := 0; i < len(categoryValueArray); i++ {
		concat := categoryArray[i] + ":" + categoryValueArray[i] + ",\n"
		extraInfoString = extraInfoString + concat
	}
	fetchedInformation = Model.Product{
		Title:        title,
		Description:  description,
		Rating:       rating,
		Price:        price,
		ExtraDetails: extraInfoString,
	}
	return fetchedInformation, nil
}
