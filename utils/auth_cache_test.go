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

func TestLoggerAttrsFromUser(t *testing.T) {
	userID := primitive.NewObjectID()
	attrs := LoggerAttrsFromUser(model.User{ID: userID, Username: "alice", DeviceID: "device-42"}, "")

	if len(attrs) != 2 {
		t.Fatalf("expected 2 log attrs, got %d", len(attrs))
	}

	if got := attrs[0].Key; got != "userId" {
		t.Fatalf("expected userId key, got %s", got)
	}
	if got := attrs[0].Value.AsString(); got != userID.Hex() {
		t.Fatalf("expected user id %s, got %s", userID.Hex(), got)
	}

	if got := attrs[1].Key; got != "deviceId" {
		t.Fatalf("expected deviceId key, got %s", got)
	}
	if got := attrs[1].Value.AsString(); got != "device-42" {
		t.Fatalf("expected device id device-42, got %s", got)
	}
}
