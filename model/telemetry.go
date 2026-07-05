package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Telemetry struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Sensor readings
	Temperature float64 `json:"temperature" bson:"temperature"`
	Humidity    float64 `json:"humidity" bson:"humidity"`
	Motor       int     `json:"motor" bson:"motor"`
	Fan         int     `json:"fan" bson:"fan"`
	Heater      int     `json:"heater" bson:"heater"`
	Spare       int     `json:"spare" bson:"spare"`

	// Device status
	GprsStatus    string `json:"gprsStatus" bson:"gprsStatus"`
	DeviceID      string `json:"deviceId" bson:"deviceId"`
	Uptime        int    `json:"uptime" bson:"uptime"`
	PowerOn       int    `json:"powerOn" bson:"powerOn"`
	SensorFailure bool   `json:"sensorFailure" bson:"sensorFailure"`

	// Mode configuration
	Mode string `json:"mode" bson:"mode"`

	// Thresholds
	FanTempMin     float64 `json:"fanTempMin" bson:"fanTempMin"`
	FanTempMax     float64 `json:"fanTempMax" bson:"fanTempMax"`
	MotorHumMin    float64 `json:"motorHumMin" bson:"motorHumMin"`
	MotorHumMax    float64 `json:"motorHumMax" bson:"motorHumMax"`
	HeaterTempMin  float64 `json:"heaterTempMin" bson:"heaterTempMin"`
	HeaterTempMax  float64 `json:"heaterTempMax" bson:"heaterTempMax"`
	FanOnDuration  float64 `json:"fanOnDuration" bson:"fanOnDuration"`
	FanOffDuration float64 `json:"fanOffDuration" bson:"fanOffDuration"`

	// Stage information
	ActiveStage            int             `json:"activeStage" bson:"activeStage"`
	StageTemp              float64         `json:"stageTemp" bson:"stageTemp"`
	StageHum               float64         `json:"stageHum" bson:"stageHum"`
	StageElapsedHours      float64         `json:"stageElapsedHours" bson:"stageElapsedHours"`
	StageRemainingHours    float64         `json:"stageRemainingHours" bson:"stageRemainingHours"`
	Stages                 []StageSettings `json:"stages" bson:"stages"`
	LastTelemetryTimestamp int64           `json:"lastTelemetryTimestamp" bson:"lastTelemetryTimestamp"`
	ProductionCompleted    bool            `json:"productionCompleted" bson:"productionCompleted"`
	UpdatedAt              time.Time       `json:"updatedAt" bson:"updatedAt"`
}

type StageSettings struct {
	TempSetpoint  float64 `json:"tempSetpoint" bson:"tempSetpoint"`
	HumSetpoint   float64 `json:"humSetpoint" bson:"humSetpoint"`
	DurationHours int     `json:"durationHours" bson:"durationHours"`
	FanTempMin    float64 `json:"fanTempMin" bson:"fanTempMin"`
	FanTempMax    float64 `json:"fanTempMax" bson:"fanTempMax"`
	MotorHumMin   float64 `json:"motorHumMin" bson:"motorHumMin"`
	MotorHumMax   float64 `json:"motorHumMax" bson:"motorHumMax"`
	HeaterTempMin float64 `json:"heaterTempMin" bson:"heaterTempMin"`
	HeaterTempMax float64 `json:"heaterTempMax" bson:"heaterTempMax"`
}
