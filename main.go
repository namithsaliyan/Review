package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Review represents a review submitted by a user
type Review struct {
	Name   string `json:"name"`
	Review string `json:"review"`
}

// Slice to store reviews
var reviews []Review

// Mutex to synchronize access to the reviews slice
var mutex = &sync.Mutex{}

func main() {
	http.HandleFunc("/reviews", reviewsHandler)

	fmt.Println("Server is listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// reviewsHandler handles both POST and GET requests for reviews
func reviewsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostReview(w, r)
	case http.MethodGet:
		handleGetReviews(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePostReview handles the submission of a new review
func handlePostReview(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var newReview Review
	if err := json.NewDecoder(r.Body).Decode(&newReview); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Lock the mutex before modifying the slice
	mutex.Lock()
	reviews = append(reviews, newReview)
	mutex.Unlock()

	// Respond with success
	response := map[string]bool{"success": true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetReviews handles fetching all submitted reviews
func handleGetReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Lock the mutex before reading the slice
	mutex.Lock()
	defer mutex.Unlock()

	json.NewEncoder(w).Encode(reviews)
}
