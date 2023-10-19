package utils

import (
	"errors"
	"os"
)

func ValidateToken(token string, tokenType string) (bool, error) {
	var validToken string
	if tokenType == "read" {
		validToken = os.Getenv("READ_AUTH_TOKEN")
	} else if tokenType == "write" {
		validToken = os.Getenv("WRITE_AUTH_TOKEN")
	} else {
		return false, errors.New("invalid token type")
	}
	if token != validToken {
		return false, nil
	}
	return true, nil
}
