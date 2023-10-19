// handlers.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SmallcaseCurl struct {
	SmallCaseId   string `bson:"smallcase_id,omitempty" json:"smallcase_id,omitempty"`
	EncryptedCurl string `bson:"encrypted_curl,omitempty"`
}

func getConstituents(w http.ResponseWriter, r *http.Request) {
	// Get query parameters

	// Find person in database
	var curlDetails SmallcaseCurl
	filter := bson.M{"smallcase_id": "CMMO_0001"}
	err := collection.FindOne(context.Background(), filter).Decode(&curlDetails)
	if err != nil {
		http.Error(w, "Smallcase not found", http.StatusNotFound)
		return
	}

	// Return person as JSON
	curl, err := DecryptText(curlDetails.EncryptedCurl)
	if err != nil {
		http.Error(w, "Error decrypting text " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cmd := exec.Command(curl)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error fetching constituents " + err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(output)

	json.NewEncoder(w).Encode(curl)
}

func setConstituents(w http.ResponseWriter, r *http.Request) {
	// Get request body
	var requestBody struct {
		Curl string `json:"curl"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Encrypt text
	encryptedText, err := EncryptText(requestBody.Curl)
	if err != nil {
		http.Error(w, "Error encrypting text " + err.Error(), http.StatusInternalServerError)
		return
	}

	// Update database
	filter := bson.M{"smallcase_id": "CMMO_0001"}
	update := bson.M{
		"$set": bson.M{
			"encrypted_curl": encryptedText,
			"updated_at":     time.Now(),
		},
	}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Error updating database", http.StatusInternalServerError)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Curl updated successfully"))
}
