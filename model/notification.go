package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type      string             `json:"type" bson:"type"`
	DeviceID  string             `json:"deviceId" bson:"deviceId"`
	Title     string             `json:"title" bson:"title"`
	Body      string             `json:"body" bson:"body"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
