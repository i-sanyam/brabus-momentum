// handlers.go

package main

import (
	"context"
	"encoding/json"
	//"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SmallcaseCurl struct {
	SmallCaseId   string `bson:"smallcase_id,omitempty" json:"smallcase_id,omitempty"`
	EncryptedCurl string `bson:"encrypted_curl,omitempty"`
}

type SmallcaseAPIResponse struct {
	Data struct {
		Constituents []interface{} `json:"constituents"`
	} `json:"data"`
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
	//fmt.Println(curl)

	w.Header().Set("Content-Type", "application/json")

	cmd := exec.Command("bash", "-c", curl)

	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error fetching constituents "+err.Error(), http.StatusInternalServerError)
		return
	}

	var response SmallcaseAPIResponse
	err = json.Unmarshal(output, &response)
	if err != nil {
		http.Error(w, "Error decoding response JSON "+err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Println(response)

	constituents := response.Data.Constituents

	json.NewEncoder(w).Encode(constituents)
}

func setConstituents(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body "+err.Error(), http.StatusInternalServerError)
		return
	}

	encryptedText, err := EncryptText(string(body))
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
