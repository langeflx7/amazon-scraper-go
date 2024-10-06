package main

import (
	"fmt"
	"github.com/langeflx7/amazonscrapergo/src/Model"
	"log"
)

var (
	productInfo Model.Product
	err         error
)

func main() {
	//sampleAsin := "B0B4T2BBHR"
	//
	//// Pass ASIN Number to function: 'FetchProductInfo'
	//productInfo.Title, productInfo.Description, productInfo.Rating, productInfo.Price, err = FetchProductInfo(sampleAsin)
	//if err != nil {
	//	log.Fatalf("Fehler beim Abrufen der Produktinformationen: %v", err)
	//}
	//fmt.Println("Title: " + productInfo.Title + "\n" + "Description: " + productInfo.Description + "\n" + "Rating: " + productInfo.Rating + "\n" + "Price: " + productInfo.Price)
	//

	// Detaillierte Bewertungen abrufen
	detailedReviews, err := FetchProductReviews(reviewsURL)
	if err != nil {
		log.Fatalf("Fehler beim Abrufen der Bewertungen: %v", err)
	}

	// Ausgabe der Bewertungen
	fmt.Println("\n--- Detaillierte Bewertungen ---")
	for _, review := range detailedReviews {
		fmt.Println(review)
	}
}
