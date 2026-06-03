package model

type Notification struct {
	Type      string `json:"type" bson:"type"`
	DeviceID  string `json:"deviceId" bson:"deviceId"`
	Title     string `json:"title" bson:"title"`
	Body      string `json:"body" bson:"body"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
}
