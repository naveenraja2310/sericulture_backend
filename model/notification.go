package model

type Notification struct {
	Type      string `json:"type" bson:"type"`
	DeviceID  string `json:"deviceId" bson:"deviceId"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
}
