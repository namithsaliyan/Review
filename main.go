package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// Review represents a review structure
type Review struct {
	Name   string `json:"name"`
	Review string `json:"review"`
}

// Global variables for storing reviews in-memory and SQLite database
var (
	reviews []Review
	mutex   = &sync.Mutex{}
	db      *sql.DB
)

func main() {
	// Initialize the database
	if err := initDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Load reviews from the database into memory on startup
	loadReviewsFromDB()

	// Set up HTTP routes
	http.HandleFunc("/reviews", handleReviews)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// initDB initializes the SQLite database
func initDB() error {
	var err error

	// Check if the database file exists
	dbFile := "./reviews.db"
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Println("Database file does not exist. Creating a new one.")
		// Create the database file
		file, err := os.Create(dbFile)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	// Open the SQLite database
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open the database: %w", err)
	}

	// Create table if it does not exist
	query := `
	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		review TEXT
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// loadReviewsFromDB loads all reviews from the SQLite database into memory
func loadReviewsFromDB() {
	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT name, review FROM reviews")
	if err != nil {
		log.Printf("Failed to fetch reviews from the database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.Name, &review.Review); err != nil {
			log.Printf("Failed to scan review: %v", err)
			continue
		}
		reviews = append(reviews, review)
	}
}

// handleReviews handles both GET and POST requests for reviews
func handleReviews(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCORS(w)

	switch r.Method {
	case http.MethodGet:
		getReviews(w, r)
	case http.MethodPost:
		postReview(w, r)
	case http.MethodOptions:
		// Handle pre-flight OPTIONS request
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getReviews handles GET requests to fetch all reviews
func getReviews(w http.ResponseWriter, r *http.Request) {
	// Lock mutex to protect reviews slice from concurrent access
	mutex.Lock()
	defer mutex.Unlock()

	// Convert reviews slice to JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// postReview handles POST requests to add a new review
func postReview(w http.ResponseWriter, r *http.Request) {
	var newReview Review

	// Decode JSON request body into a Review struct
	err := json.NewDecoder(r.Body).Decode(&newReview)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Lock mutex to protect reviews slice from concurrent access
	mutex.Lock()
	reviews = append(reviews, newReview)
	mutex.Unlock()

	// Save review to the database
	if err := saveReviewToDB(newReview); err != nil {
		http.Error(w, "Failed to save review to the database", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// saveReviewToDB saves a new review to the SQLite database
func saveReviewToDB(review Review) error {
	_, err := db.Exec("INSERT INTO reviews (name, review) VALUES (?, ?)", review.Name, review.Review)
	if err != nil {
		log.Printf("Failed to save review to the database: %v", err)
		return err
	}
	return nil
}

// enableCORS sets CORS headers to allow cross-origin requests
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
