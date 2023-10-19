// handlers.go

package main

import (
	"strings"
	"io"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
	"fmt"

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
	fmt.Println(curl)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error fetching constituents " + err.Error(), http.StatusInternalServerError)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		http.Error(w, "Error decoding constituents JSON " + err.Error(), http.StatusInternalServerError)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert body to string and remove "curl" prefix
	command := strings.TrimPrefix(string(body), "curl ")

	// Encrypt text
	encryptedText, err := EncryptText(command)
	//encryptedText, err := EncryptText(string(body))
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

	// shortcircuit to return API response
	getConstituents(w, r)
}
