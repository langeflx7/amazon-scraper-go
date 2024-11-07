package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/langeflx7/amazonscrapergo/src/Model"
	"github.com/langeflx7/amazonscrapergo/src/productviewer"
	"github.com/rs/cors"
)

var (
	productInfo Model.Product
	db          *sql.DB
	err         error
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

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
	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow your frontend's origin
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	// Set up routes with gorilla mux
	r := mux.NewRouter()
	r.HandleFunc("/fetch-product", fetchProductHandler).Methods("POST")
	r.HandleFunc("/healthz", healthCheckHandler).Methods("GET") // Health check endpoint

	// Start the server with CORS middleware
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}

// fetchProductHandler handles requests to fetch product data and update the database
func fetchProductHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := checkDBConnection(); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Parse the JSON body
	var requestData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		// Return consistent error response
		response := Response{Status: "error", Error: "Invalid JSON body"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Extract product_id and url from the parsed body
	productID, ok := requestData["product_id"].(string)
	if !ok {
		// Return consistent error response
		response := Response{Status: "error", Error: "Invalid product_id format"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	url, ok := requestData["url"].(string)
	if !ok {
		// Return consistent error response
		response := Response{Status: "error", Error: "Invalid url format"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Log incoming request
	fmt.Printf("Incoming request to fetch product with product_id: %s\n", productID)

	// Extract ASIN from the provided URL
	asin := extractASIN(url)
	if asin == "" {
		// Return consistent error response
		response := Response{Status: "error", Error: "Invalid URL, ASIN not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fetch product info based on ASIN
	productInfo, err = productviewer.FetchProductInfo(asin)
	if err != nil {
		// Return consistent error response
		response := Response{Status: "error", Error: fmt.Sprintf("Failed to fetch product info: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Log fetched product data (excluding description and details)
	fmt.Println("Fetched product data:")
	fmt.Printf("Title: %s\n", productInfo.Title)
	fmt.Printf("Price: %f\n", productInfo.Price)
	fmt.Printf("Rating: %f\n", productInfo.Rating)

	// Update product_info table in MySQL database
	err = updateProductInfoInDB(productID, productInfo)
	if err != nil {
		// Return consistent error response
		response := Response{Status: "error", Error: fmt.Sprintf("Failed to update product info in database: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send success response with product data
	responseData := map[string]interface{}{
		"status": "ok",
		"data":   productInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
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
func updateProductInfoInDB(productID string, productInfo Model.Product) error {
	query := `UPDATE product_info
				SET title = ?, description = ?, price = ?, rating = ?
				WHERE product_id = ?`

	_, err := db.Exec(query, productInfo.Title, productInfo.Description, productInfo.Price, productInfo.Rating, productID)
	return err
}
