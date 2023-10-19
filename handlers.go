// handlers.go

package main

import (
	"context"
	"encoding/json"
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
	var curlDetails SmallcaseCurl
	filter := bson.M{"smallcase_id": "CMMO_0001"}
	err := collection.FindOne(context.Background(), filter).Decode(&curlDetails)
	if err != nil {
		http.Error(w, "Smallcase not found", http.StatusNotFound)
		return
	}

	curl, err := DecryptText(curlDetails.EncryptedCurl)
	if err != nil {
		http.Error(w, "Error decrypting text "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cmd := exec.Command("curl", curl)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error fetching constituents "+err.Error(), http.StatusInternalServerError)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		http.Error(w, "Error decoding constituents JSON "+err.Error(), http.StatusInternalServerError)
		return
	}

	constituents, ok := data["constituents"].([]interface{})
	if !ok {
		http.Error(w, "Error extracting constituents", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(constituents)
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
		http.Error(w, "Error encrypting text "+err.Error(), http.StatusInternalServerError)
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
