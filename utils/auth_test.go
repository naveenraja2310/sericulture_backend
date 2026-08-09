package utils

import (
	"testing"
	"time"

	"sericulture/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateAndValidateTokenForSamePassword(t *testing.T) {
	user := model.User{
		ID:        primitive.NewObjectID(),
		Username:  "alice",
		Password:  "super-secret",
		CreatedAt: time.Now(),
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT returned unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateJWT returned empty token")
	}

	if err := ValidateJWT(token, user); err != nil {
		t.Fatalf("ValidateJWT rejected a valid token: %v", err)
	}
}

func TestValidateTokenRejectsChangedPassword(t *testing.T) {
	user := model.User{
		ID:        primitive.NewObjectID(),
		Username:  "alice",
		Password:  "original-pass",
		CreatedAt: time.Now(),
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT returned unexpected error: %v", err)
	}

	updatedUser := user
	updatedUser.Password = "new-pass"

	if err := ValidateJWT(token, updatedUser); err == nil {
		t.Fatal("ValidateJWT accepted a token after password change")
	}
}
