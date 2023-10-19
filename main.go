package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	// Load ENV variables
	godotenv.Load()

	// generate new key
	//GenerateKey();

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("DATABASE_NAME")).Collection("smallcase_curls")

	// Create router
	router := mux.NewRouter()

	// Define API endpoints
	router.Handle("/constituents", authMiddleware("read")(http.HandlerFunc(getConstituents))).Methods("GET")
	router.Handle("/constituents", authMiddleware("write")(http.HandlerFunc(setConstituents))).Methods("POST")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
