package main

import (
	"fmt"
	"github.com/langeflx7/amazonscrapergo/src/Model"
	"log"
	"strconv"
)

var (
	productInfo Model.Product
	err         error
)

func main() {
	sampleAsin := "B0B4T2BBHR"

	// Pass ASIN Number to function: 'FetchProductInfo'
	productInfo, err = FetchProductInfo(sampleAsin)
	if err != nil {
		log.Fatalf("Failed to fetch product information : %v", err)
	}
	fmt.Println(productInfo.ExtraDetails)
	fmt.Println("Title: " + productInfo.Title + "\n" + "Description: " + productInfo.Description + "\n" + "Rating: " + productInfo.Rating + "\n" + "Price: " + productInfo.Price + "\n" + "Extra Information" + productInfo.ExtraDetails)

	// Review return as per count passed
	detailedReviews, err := FetchProductReviews(sampleAsin, 20)
	if err != nil {
		log.Fatalf("Failed to fetch Product reviews: %v", err)
	}
	fmt.Println("Number of review: " + strconv.Itoa(len(detailedReviews)))
	for _, review := range detailedReviews {
		fmt.Println("Review: " + review + "\n")
	}
}
