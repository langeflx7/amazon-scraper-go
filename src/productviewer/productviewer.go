package productviewer

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
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
		Price:       0.0, // Default value for price as float64
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

	// Price extraction and cleanup
	var price string
	priceClass := strings.TrimSpace(doc.Find("span.aok-offscreen").Text())
	if priceClass != "" {
		// Clean the price
		price = cleanPrice(priceClass)
	} else {
		priceFromButton := strings.TrimSpace(doc.Find("span.a-size-mini.olpWrapper").Text())
		price = priceFromButton
		if price == "" {
			doc.Find("span.a-price.a-text-price.a-size-medium").Each(func(i int, s *goquery.Selection) {
				price = "Ã¢â€šÂ¬" + strings.Split(s.Text(), "Ã¢â€šÂ¬")[1]
			})
		}
	}

	// Debug output for the price before conversion
	fmt.Printf("Extracted Price (before cleanup): %s\n", price)

	// Clean and convert the price string
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		// If price parsing fails, return 0.0
		fmt.Printf("Error parsing price: %s. Setting to 0.0\n", price)
		priceFloat = 0.0
	}

	// Final fetched product information
	fetchedInformation = Model.Product{
		Title:        title,
		Description:  description,
		Rating:       rating,
		Price:        priceFloat, // Store as float64
		ExtraDetails: "Extra info not extracted",
	}

	return fetchedInformation, nil
}

// cleanPrice entfernt alle nicht numerischen Zeichen und konvertiert den Preis zu einem float64 mit maximal zwei Dezimalstellen
func cleanPrice(price string) string {
	// Step 1: Entferne alle nicht numerischen Zeichen (auÃŸer Komma und Punkt)
	re := regexp.MustCompile(`[^\d,\.]`)           // Entferne alles, was keine Ziffer, Komma oder Punkt ist
	cleanedPrice := re.ReplaceAllString(price, "") // Entferne unerwÃ¼nschte Zeichen

	// Debug-Ausgabe: Preis nach der ersten Bereinigung
	fmt.Printf("Cleaned Price: %s\n", cleanedPrice)

	// Step 2: Ersetze das Komma durch einen Punkt, falls vorhanden
	if strings.Contains(cleanedPrice, ",") {
		// Ersetze das erste Komma mit einem Punkt fÃ¼r eine korrekte Umwandlung in float64
		cleanedPrice = strings.Replace(cleanedPrice, ",", ".", 1)
	}

	// Step 3: ÃœberprÃ¼fe, ob mehr als ein Punkt vorhanden ist
	if strings.Count(cleanedPrice, ".") > 1 {
		// Wenn mehrere Punkte vorhanden sind, behalte nur den ersten Punkt und entferne alle anderen
		splitPrice := strings.Split(cleanedPrice, ".")
		cleanedPrice = splitPrice[0] + "." + strings.Join(splitPrice[1:], "")
	}

	// Debug-Ausgabe: Preis nach der Umformatierung
	fmt.Printf("Cleaned Price After Formatting: %s\n", cleanedPrice)

	// Step 4: Konvertiere den Preis in float64 und runde auf 2 Dezimalstellen
	priceFloat, err := strconv.ParseFloat(cleanedPrice, 64)
	if err != nil {
		// Fehler beim Parsen, setze auf 0.0
		fmt.Println("Error parsing price:", cleanedPrice, "Setting to 0.0")
		return "0.0"
	}

	// Runde auf 2 Dezimalstellen
	roundedPrice := fmt.Sprintf("%.2f", priceFloat)

	// Debug-Ausgabe: Gerundeter Preis
	fmt.Printf("Rounded Price: %s\n", roundedPrice)

	// Gebe den gerundeten Preis zurÃ¼ck
	return roundedPrice
}
