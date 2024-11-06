package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/langeflx7/amazonscrapergo/src/Model"
	"github.com/langeflx7/amazonscrapergo/src/productviewer"
)

var (
	productInfo Model.Product
	db          *sql.DB
	err         error
)

// init establishes the initial database connection with a configured connection pool
func init() {
	reconnectDB()
	// Configure database connection pooling settings
	db.SetMaxOpenConns(10)           // Maximum number of open connections
	db.SetMaxIdleConns(5)            // Maximum number of idle connections
	db.SetConnMaxLifetime(time.Hour) // Recycle connections periodically
}

// reconnectDB attempts to establish a new database connection
func reconnectDB() error {
	dsn := "dev:WebFlinkSQLDev@tcp(localhost:3306)/proda"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return err
	}
	return db.Ping()
}

// checkDBConnection ensures the database connection is live, reconnecting if necessary
func checkDBConnection() error {
	if err := db.Ping(); err != nil {
		log.Println("Lost database connection, attempting to reconnect...")
		return reconnectDB()
	}
	return nil
}

func main() {
	// HTTP routes and server setup
	http.HandleFunc("/fetch-product", fetchProductHandler)
	http.HandleFunc("/healthz", healthCheckHandler) // Health check endpoint
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// fetchProductHandler handles requests to fetch product data and update the database
func fetchProductHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := checkDBConnection(); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Extract product_id and url from query parameters
	productIDStr := r.URL.Query().Get("product_id")
	url := r.URL.Query().Get("url")

	// Convert product_id to an integer
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product_id", http.StatusBadRequest)
		return
	}

	// Extract ASIN from the provided URL
	asin := extractASIN(url)
	if asin == "" {
		http.Error(w, "Invalid URL, ASIN not found", http.StatusBadRequest)
		return
	}

	// Fetch product info based on ASIN
	productInfo, err = productviewer.FetchProductInfo(asin)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch product info: %v", err), http.StatusInternalServerError)
		return
	}

	// Update product_info table in MySQL database
	err = updateProductInfoInDB(productID, productInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update product info in database: %v", err), http.StatusInternalServerError)
		return
	}

	// Send product data back as JSON response
	responseData, err := json.Marshal(productInfo)
	if err != nil {
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

// healthCheckHandler provides a health check endpoint to verify the server and database status
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "Database not reachable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// extractASIN extracts the ASIN from a given URL
func extractASIN(url string) string {
	re := regexp.MustCompile(`/([A-Z0-9]{10})(?:[/?]|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// updateProductInfoInDB updates the product_info table in MySQL with product data
func updateProductInfoInDB(productID int, productInfo Model.Product) error {
	query := `UPDATE product_info
				SET title = ?, description = ?, price = ?, rating = ?
				WHERE product_id = ?`

	_, err := db.Exec(query, productInfo.Title, productInfo.Description, productInfo.Price, productInfo.Rating, productID)
	return err
}
