package model

import (
	"encoding/json"
	"strings"
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
	GprsStatus         string  `json:"gprsStatus" bson:"gprsStatus"`
	DeviceID           string  `json:"deviceId" bson:"deviceId"`
	Uptime             int     `json:"uptime" bson:"uptime"`
	PowerOn            int     `json:"powerOn" bson:"powerOn"`
	SensorFailure      bool    `json:"sensorFailure" bson:"sensorFailure"`
	Timer              int     `json:"timer" bson:"timer"`
	DehumidifierHum    float64 `json:"dehumidifierHum" bson:"dehumidifierHum"`
	DehumidifierActive bool    `json:"dehumidifierActive" bson:"dehumidifierActive"`
	RelaysOn           int     `json:"relaysOn" bson:"relaysOn"`

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
	StageDurationHours     int             `json:"stageDurationHours" bson:"stageDurationHours"`
	StageElapsedHours      float64         `json:"stageElapsedHours" bson:"stageElapsedHours"`
	StageRemainingHours    float64         `json:"stageRemainingHours" bson:"stageRemainingHours"`
	Stages                 []StageSettings `json:"stages" bson:"stages"`
	LastTelemetryTimestamp int64           `json:"lastTelemetryTimestamp" bson:"lastTelemetryTimestamp"`
	ProductionCompleted    bool            `json:"productionCompleted" bson:"productionCompleted"`
	FirmwareVersion        string          `json:"firmwareVersion" bson:"firmwareVersion"`
	SystemEnabled          bool            `json:"systemEnabled" bson:"systemEnabled"`
	UpdatedAt              time.Time       `json:"updatedAt" bson:"updatedAt"`
}

type StageSettings struct {
	FanTempMin      float64 `json:"fanTempMin" bson:"fanTempMin"`
	FanTempMax      float64 `json:"fanTempMax" bson:"fanTempMax"`
	MotorHumMin     float64 `json:"motorHumMin" bson:"motorHumMin"`
	MotorHumMax     float64 `json:"motorHumMax" bson:"motorHumMax"`
	HeaterTempMin   float64 `json:"heaterTempMin" bson:"heaterTempMin"`
	HeaterTempMax   float64 `json:"heaterTempMax" bson:"heaterTempMax"`
	DurationHours   int     `json:"durationHours" bson:"durationHours"`
	FanOnDuration   float64 `json:"fanOnDuration" bson:"fanOnDuration"`   // NEW
	FanOffDuration  float64 `json:"fanOffDuration" bson:"fanOffDuration"` // NEW
	DehumidifierHum float64 `json:"dehumidifierHum" bson:"dehumidifierHum"`
}

// UnmarshalJSON accepts both full-size firmware keys and compact firmware keys.
// Compact keys are checked first, so they win when both names are present.
func (telemetry *Telemetry) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	decode := func(target interface{}, keys ...string) error {
		raw, ok := preferredField(fields, keys...)
		if !ok {
			return nil
		}
		return json.Unmarshal(raw, target)
	}
	decodeStringOrNumber := func(target *string, keys ...string) error {
		raw, ok := preferredField(fields, keys...)
		if !ok {
			return nil
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil {
			*target = value
			return nil
		}
		var number json.Number
		if err := json.Unmarshal(raw, &number); err != nil {
			return err
		}
		*target = number.String()
		return nil
	}
	decodeBool := func(target *bool, keys ...string) error {
		raw, ok := preferredField(fields, keys...)
		if !ok {
			return nil
		}
		var value bool
		if err := json.Unmarshal(raw, &value); err == nil {
			*target = value
			return nil
		}
		var number int
		if err := json.Unmarshal(raw, &number); err != nil {
			return err
		}
		*target = number != 0
		return nil
	}

	if err := decode(&telemetry.Temperature, "t", "temperature"); err != nil {
		return err
	}
	if err := decode(&telemetry.Humidity, "h", "humidity"); err != nil {
		return err
	}
	if err := decode(&telemetry.Motor, "m", "motor"); err != nil {
		return err
	}
	if err := decode(&telemetry.Fan, "f", "fan"); err != nil {
		return err
	}
	if err := decode(&telemetry.Heater, "ht", "heater"); err != nil {
		return err
	}
	if err := decode(&telemetry.Spare, "sp", "spare"); err != nil {
		return err
	}
	if err := decodeStringOrNumber(&telemetry.GprsStatus, "gprs", "gprsStatus"); err != nil {
		return err
	}
	if err := decode(&telemetry.DeviceID, "id", "deviceId"); err != nil {
		return err
	}
	if err := decode(&telemetry.Uptime, "up", "uptime"); err != nil {
		return err
	}
	if err := decode(&telemetry.PowerOn, "relaysOn", "powerOn"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.SensorFailure, "senFail", "sensorFailure"); err != nil {
		return err
	}
	if err := decode(&telemetry.Timer, "tim", "timer"); err != nil {
		return err
	}
	if err := decode(&telemetry.Mode, "mode"); err != nil {
		return err
	}
	telemetry.Mode = NormalizeMode(telemetry.Mode)
	if err := decode(&telemetry.FanTempMin, "ftMin", "fanTempMin"); err != nil {
		return err
	}
	if err := decode(&telemetry.FanTempMax, "ftMax", "fanTempMax"); err != nil {
		return err
	}
	if err := decode(&telemetry.MotorHumMin, "mhMin", "motorHumMin"); err != nil {
		return err
	}
	if err := decode(&telemetry.MotorHumMax, "mhMax", "motorHumMax"); err != nil {
		return err
	}
	if err := decode(&telemetry.HeaterTempMin, "htMin", "heaterTempMin"); err != nil {
		return err
	}
	if err := decode(&telemetry.HeaterTempMax, "htMax", "heaterTempMax"); err != nil {
		return err
	}
	if err := decode(&telemetry.FanOnDuration, "fon", "fanOnDuration"); err != nil {
		return err
	}
	if err := decode(&telemetry.FanOffDuration, "foff", "fanOffDuration"); err != nil {
		return err
	}
	if err := decode(&telemetry.ActiveStage, "stg", "activeStage"); err != nil {
		return err
	}
	if err := decode(&telemetry.StageDurationHours, "dur", "stageDurationHours"); err != nil {
		return err
	}
	if err := decode(&telemetry.StageElapsedHours, "elap", "stageElapsedHours"); err != nil {
		return err
	}
	if err := decode(&telemetry.StageRemainingHours, "rem", "stageRemainingHours"); err != nil {
		return err
	}
	if err := decode(&telemetry.Stages, "stages"); err != nil {
		return err
	}
	if err := decode(&telemetry.LastTelemetryTimestamp, "lastTelemetryTimestamp"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.ProductionCompleted, "prodDone", "productionCompleted"); err != nil {
		return err
	}
	if err := decode(&telemetry.FirmwareVersion, "ver", "firmwareVersion"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.SystemEnabled, "en", "systemEnabled"); err != nil {
		return err
	}
	if err := decode(&telemetry.DehumidifierHum, "dehum", "dehumidifierHum"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.DehumidifierActive, "dehAct", "dehumidifierActive"); err != nil {
		return err
	}
	if err := decode(&telemetry.RelaysOn, "relaysOn", "powerOn"); err != nil {
		return err
	}

	return decode(&telemetry.UpdatedAt, "updatedAt")
}

func NormalizeMode(mode string) string {
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case "A", "AUTO":
		return "AUTO"
	case "M", "MANUAL":
		return "MANUAL"
	default:
		return mode
	}
}

func (settings *StageSettings) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decode := func(target interface{}, keys ...string) error {
		raw, ok := preferredField(fields, keys...)
		if !ok {
			return nil
		}
		return json.Unmarshal(raw, target)
	}

	if err := decode(&settings.FanTempMin, "ftMin", "fanTempMin"); err != nil {
		return err
	}
	if err := decode(&settings.FanTempMax, "ftMax", "fanTempMax"); err != nil {
		return err
	}
	if err := decode(&settings.MotorHumMin, "mhMin", "motorHumMin"); err != nil {
		return err
	}
	if err := decode(&settings.MotorHumMax, "mhMax", "motorHumMax"); err != nil {
		return err
	}
	if err := decode(&settings.HeaterTempMin, "htMin", "heaterTempMin"); err != nil {
		return err
	}
	if err := decode(&settings.HeaterTempMax, "htMax", "heaterTempMax"); err != nil {
		return err
	}
	if err := decode(&settings.DurationHours, "dur", "durationHours"); err != nil {
		return err
	}
	if err := decode(&settings.FanOnDuration, "fon", "fanOnDuration"); err != nil {
		return err
	}
	if err := decode(&settings.FanOffDuration, "foff", "fanOffDuration"); err != nil {
		return err
	}
	return decode(&settings.DehumidifierHum, "dehum", "dehumidifierHum")
}

func preferredField(fields map[string]json.RawMessage, keys ...string) (json.RawMessage, bool) {
	for _, key := range keys {
		if raw, ok := fields[key]; ok && string(raw) != "null" {
			return raw, true
		}
	}
	return nil, false
}
