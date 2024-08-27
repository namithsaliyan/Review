package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Review represents a review structure
type Review struct {
	Name   string `json:"name"`
	Review string `json:"review"`
}

// Global variables for storing reviews in-memory
var reviews []Review
var mutex = &sync.Mutex{}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/reviews", handleReviews)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// enableCORS sets CORS headers to allow cross-origin requests
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
