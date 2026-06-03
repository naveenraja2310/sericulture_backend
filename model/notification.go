package model

type Notification struct {
	Type      string `json:"type" bson:"type"`
	DeviceID  string `json:"deviceId" bson:"deviceId"`
	Title     string `json:"title" bson:"title"`
	Message   string `json:"message" bson:"message"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
}
