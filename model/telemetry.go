package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Telemetry struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Temperature float64            `json:"temperature"`
	Humidity    float64            `json:"humidity"`

	Motor  int `json:"motor"`
	Fan    int `json:"fan"`
	Heater int `json:"heater"`
	Spare  int `json:"spare"`

	Mode string `json:"mode"`

	TempThreshold float64 `json:"tempThreshold"`
	HumThreshold  float64 `json:"humThreshold"`

	FanCycleTime int `json:"fanCycleTime"`

	Uptime int `json:"uptime"`

	GprsStatus string `json:"gprsStatus"`

	DeviceID string `json:"deviceId"`

	PowerOn int `json:"powerOn"`

	UpdatedAt time.Time `json:"updatedAt"`
}
