package main

import (
	"fmt"
	"log"
)

func main() {
	productURL := "https://www.amazon.de/Devoko-H%C3%B6henverstellbarer-Computertisch-2-Funktions-Memory-Ergonomischer/dp/B0CXY2M22T/"
	reviewsURL := "https://www.amazon.de/Devoko-H%C3%B6henverstellbarer-Computertisch-2-Funktions-Memory-Ergonomischer/product-reviews/B0CXY2M22T/ref=cm_cr_getr_d_paging_btm_prev_1?ie=UTF8&reviewerType=all_reviews"

	// Produktinformationen abrufen
	title, description, reviewsSummary, err := FetchProductInfo(productURL)
	if err != nil {
		log.Fatalf("Fehler beim Abrufen der Produktinformationen: %v", err)
	}

	// Ausgabe der Produktinformationen
	fmt.Println("Produkt-Titel:", title)
	fmt.Println("Produkt-Beschreibung:", description)
	fmt.Println("Zusammenfassung der Bewertungen:", reviewsSummary)

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
