package utils

import (
	"testing"

	"sericulture/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestClearUserCache(t *testing.T) {
	authUserCache.Set("user-123", model.User{ID: primitive.NewObjectID(), Username: "test"})
	if _, ok := authUserCache.Get("user-123"); !ok {
		t.Fatal("expected user in cache before clear")
	}

	ClearUserCache("user-123")

	if _, ok := authUserCache.Get("user-123"); ok {
		t.Fatal("expected user to be cleared from cache")
	}
}
