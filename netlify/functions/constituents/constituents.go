// handlers.go

package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/i-sanyam/brabus-momentum/netlify/utils"

	"github.com/joho/godotenv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
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

type FilteredConstituent struct {
	NseTicker string  `json:"nseTicker"`
	Weight    float64 `json:"weight"`
}

var collection *mongo.Collection

func init() {
	// Load ENV variables
	godotenv.Load()
	collection = utils.GetMongoCollection()
}

func getMappedConstituents(constituents []interface{}) []FilteredConstituent {
	var filteredConstituents []FilteredConstituent

	// Iterate through the response and extract the required fields
	for _, constituent := range constituents {
		if constituentMap, ok := constituent.(map[string]interface{}); ok {
			weight, _ := constituentMap["weight"].(float64)
			if sidInfo, sidInfoExists := constituentMap["sidInfo"].(map[string]interface{}); sidInfoExists {
				nseTicker, _ := sidInfo["nseTicker"].(string)
				filteredConstituent := FilteredConstituent{
					NseTicker: nseTicker,
					Weight:    weight,
				}
				filteredConstituents = append(filteredConstituents, filteredConstituent)
			}
		}
	}
	return filteredConstituents
}

func getConstituents(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var curlDetails SmallcaseCurl
	filter := bson.M{"smallcase_id": "CMMO_0001"}
	err := collection.FindOne(context.Background(), filter).Decode(&curlDetails)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Smallcase not found", StatusCode: 404}, nil
	}

	curl, err := utils.DecryptText(curlDetails.EncryptedCurl)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error decrypting text " + err.Error(), StatusCode: 500}, nil
	}

	cmd := exec.Command("bash", "-c", curl)

	output, err := cmd.Output()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error fetching constituents " + err.Error(), StatusCode: 500}, nil
	}

	var response SmallcaseAPIResponse
	err = json.Unmarshal(output, &response)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error decoding response JSON " + err.Error(), StatusCode: 500}, nil
	}

	body, err := json.Marshal(getMappedConstituents(response.Data.Constituents))
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error encoding response JSON " + err.Error(), StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func setConstituents(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := request.Body

	encryptedText, err := utils.EncryptText(body)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error encrypting text " + err.Error(), StatusCode: 500}, nil
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
		return events.APIGatewayProxyResponse{Body: "Error updating database", StatusCode: 500}, nil
	}

	// shortcircuit to return API response
	return getConstituents(ctx, request)
}

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		authToken := request.Headers["authorization"]

		switch request.HTTPMethod {
		case "GET":
			isValid, _ := utils.ValidateToken(authToken, "read")
			if !isValid {
				return events.APIGatewayProxyResponse{Body: "Unauthorized", StatusCode: 401}, nil
			}
			return getConstituents(ctx, request)
		case "POST":
			isValid, _ := utils.ValidateToken(authToken, "write")
			if !isValid {
				return events.APIGatewayProxyResponse{Body: "Unauthorized", StatusCode: 401}, nil
			}
			return setConstituents(ctx, request)
		default:
			return events.APIGatewayProxyResponse{Body: "Method not allowed", StatusCode: 405}, nil
		}
	})
}