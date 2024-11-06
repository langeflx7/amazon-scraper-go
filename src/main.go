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
	"github.com/rs/cors"
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
	// CORS-Handler setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allow every origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// HTTP-Routen einrichten
	http.HandleFunc("/fetch-product", fetchProductHandler)
	http.HandleFunc("/healthz", healthCheckHandler) // Health check endpoint

	// CORS auf alle Routen anwenden
	handlerWithCORS := c.Handler(http.DefaultServeMux)

	// Starte den Server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", handlerWithCORS))
}

// fetchProductHandler handles requests to fetch product data and update the database
func fetchProductHandler(w http.ResponseWriter, r *http.Request) {
	// Setze den Content-Type auf application/json
	w.Header().Set("Content-Type", "application/json")

	// Überprüfe die Datenbankverbindung
	if err := checkDBConnection(); err != nil {
		http.Error(w, `{"status": "error", "message": "Database connection error"}`, http.StatusInternalServerError)
		return
	}

	// Extrahiere product_id und url aus den Anfrageparametern
	productIDStr := r.URL.Query().Get("product_id")
	url := r.URL.Query().Get("url")

	// Versuche, die product_id zu einer Ganzzahl zu konvertieren
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, `{"status": "error", "message": "Invalid product_id"}`, http.StatusBadRequest)
		return
	}

	// Extrahiere ASIN aus der URL
	asin := extractASIN(url)
	if asin == "" {
		http.Error(w, `{"status": "error", "message": "Invalid URL, ASIN not found"}`, http.StatusBadRequest)
		return
	}

	// Hole die Produktinformationen basierend auf der ASIN
	productInfo, err := productviewer.FetchProductInfo(asin)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"status": "error", "message": "Failed to fetch product info: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update die Produktinformationen in der MySQL-Datenbank
	err = updateProductInfoInDB(productID, productInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"status": "error", "message": "Failed to update product info in database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Sende die Produktinformationen als JSON-Antwort zurück
	responseData, err := json.Marshal(map[string]interface{}{
		"status":  "ok",
		"product": productInfo,
	})
	if err != nil {
		http.Error(w, `{"status": "error", "message": "Failed to encode response data"}`, http.StatusInternalServerError)
		return
	}
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
