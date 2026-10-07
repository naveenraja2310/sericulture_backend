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
	RelaysActive       int     `json:"relaysActive" bson:"relaysActive"`
	SystemStatus       string  `json:"systemStatus" bson:"systemStatus"`

	// Air quality readings
	CO2       float64 `json:"co2" bson:"co2"`
	CO2Raw    float64 `json:"co2Raw" bson:"co2Raw"`
	CO2Pct    float64 `json:"co2Pct" bson:"co2Pct"`
	CO2Status string  `json:"co2Status" bson:"co2Status"`
	CO2Warn   bool    `json:"co2Warn" bson:"co2Warn"`
	NH3       float64 `json:"nh3" bson:"nh3"`
	NH3Raw    float64 `json:"nh3Raw" bson:"nh3Raw"`
	NH3Pct    float64 `json:"nh3Pct" bson:"nh3Pct"`
	NH3Rs     float64 `json:"nh3Rs" bson:"nh3Rs"`
	NH3Ratio  float64 `json:"nh3Ratio" bson:"nh3Ratio"`
	NH3Warn   bool    `json:"nh3Warn" bson:"nh3Warn"`

	// Sensor offsets
	TempOffset float64 `json:"tempOffset" bson:"tempOffset"`
	HumOffset  float64 `json:"humOffset" bson:"humOffset"`
	CO2Offset  float64 `json:"co2Offset" bson:"co2Offset"`
	NH3Offset  float64 `json:"nh3Offset" bson:"nh3Offset"`

	// Air quality thresholds and fan overrides
	CO2WarnPPM    float64 `json:"co2WarnPPM" bson:"co2WarnPPM"`
	NH3WarnPPM    float64 `json:"nh3WarnPPM" bson:"nh3WarnPPM"`
	CO2MinPct     float64 `json:"co2MinPct" bson:"co2MinPct"`
	CO2MaxPct     float64 `json:"co2MaxPct" bson:"co2MaxPct"`
	NH3MinPct     float64 `json:"nh3MinPct" bson:"nh3MinPct"`
	NH3MaxPct     float64 `json:"nh3MaxPct" bson:"nh3MaxPct"`
	CO2FanOverride bool   `json:"co2FanOverride" bson:"co2FanOverride"`
	NH3FanOverride bool   `json:"nh3FanOverride" bson:"nh3FanOverride"`
	AirOverride    bool   `json:"airOverride" bson:"airOverride"`

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
	StageNumber            int             `json:"stageNumber" bson:"stageNumber"`
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
	CO2WarnPPM      float64 `json:"co2WarnPPM" bson:"co2WarnPPM"`
	NH3WarnPPM      float64 `json:"nh3WarnPPM" bson:"nh3WarnPPM"`
	CO2MinPct       float64 `json:"co2MinPct" bson:"co2MinPct"`
	CO2MaxPct       float64 `json:"co2MaxPct" bson:"co2MaxPct"`
	NH3MinPct       float64 `json:"nh3MinPct" bson:"nh3MinPct"`
	NH3MaxPct       float64 `json:"nh3MaxPct" bson:"nh3MaxPct"`
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
	telemetry.GprsStatus = NormalizeGprsStatus(telemetry.GprsStatus)
	if err := decode(&telemetry.DeviceID, "id", "deviceId"); err != nil {
		return err
	}
	if err := decode(&telemetry.Uptime, "up", "uptime"); err != nil {
		return err
	}
	if err := decode(&telemetry.PowerOn, "po", "powerOn"); err != nil {
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
	if err := decode(&telemetry.ActiveStage, "stg", "activeStage", "stageNumber"); err != nil {
		return err
	}
	if err := decode(&telemetry.StageNumber, "stg", "stageNumber", "activeStage"); err != nil {
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
	if err := decode(&telemetry.Stages, "s", "stages"); err != nil {
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
	if err := decode(&telemetry.SystemStatus, "systemStatus"); err != nil {
		return err
	}
	telemetry.SystemStatus = NormalizeSystemStatus(telemetry.SystemStatus)
	if telemetry.SystemStatus == "" && preferredKey(fields, "en", "systemEnabled") {
		if telemetry.SystemEnabled {
			telemetry.SystemStatus = "ENABLED"
		} else {
			telemetry.SystemStatus = "DISABLED"
		}
	}
	if err := decode(&telemetry.DehumidifierHum, "dehum", "dehumidifierHum"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.DehumidifierActive, "dehAct", "dehumidifierActive"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2, "co2"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2Raw, "co2Raw"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2Pct, "co2Pct"); err != nil {
		return err
	}
	if err := decodeStringOrNumber(&telemetry.CO2Status, "co2st", "co2Status"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.CO2Warn, "co2Warn"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3, "nh3"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3Raw, "nh3Raw"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3Pct, "nh3Pct"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3Rs, "nh3Rs"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3Ratio, "nh3Ratio"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.NH3Warn, "nh3Warn"); err != nil {
		return err
	}
	if err := decode(&telemetry.TempOffset, "tOff", "tempOffset"); err != nil {
		return err
	}
	if err := decode(&telemetry.HumOffset, "hOff", "humOffset"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2Offset, "co2Off", "co2Offset"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3Offset, "nh3Off", "nh3Offset"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2WarnPPM, "co2Set", "co2WarnPPM"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3WarnPPM, "nh3Set", "nh3WarnPPM"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2MinPct, "co2Min", "co2MinPct"); err != nil {
		return err
	}
	if err := decode(&telemetry.CO2MaxPct, "co2Max", "co2MaxPct"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3MinPct, "nh3Min", "nh3MinPct"); err != nil {
		return err
	}
	if err := decode(&telemetry.NH3MaxPct, "nh3Max", "nh3MaxPct"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.CO2FanOverride, "co2Ovr", "co2FanOverride"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.NH3FanOverride, "nh3Ovr", "nh3FanOverride"); err != nil {
		return err
	}
	if err := decodeBool(&telemetry.AirOverride, "airOvr", "airOverride"); err != nil {
		return err
	}
	if err := decode(&telemetry.RelaysOn, "relaysOn", "powerOn"); err != nil {
		return err
	}
	if err := decode(&telemetry.RelaysActive, "relaysActive", "relaysOn", "powerOn"); err != nil {
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

func NormalizeGprsStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "1", "TRUE", "CONNECTED", "ONLINE":
		return "CONNECTED"
	case "0", "FALSE", "DISCONNECTED", "OFFLINE":
		return "DISCONNECTED"
	default:
		return status
	}
}

func NormalizeSystemStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "1", "TRUE", "ENABLED", "ON":
		return "ENABLED"
	case "0", "FALSE", "DISABLED", "OFF":
		return "DISABLED"
	default:
		return status
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

	if err := decode(&settings.FanTempMin, "fm", "ftMin", "fanTempMin"); err != nil {
		return err
	}
	if err := decode(&settings.FanTempMax, "fx", "ftMax", "fanTempMax"); err != nil {
		return err
	}
	if err := decode(&settings.MotorHumMin, "mm", "mhMin", "motorHumMin"); err != nil {
		return err
	}
	if err := decode(&settings.MotorHumMax, "mx", "mhMax", "motorHumMax"); err != nil {
		return err
	}
	if err := decode(&settings.HeaterTempMin, "hm", "htMin", "heaterTempMin"); err != nil {
		return err
	}
	if err := decode(&settings.HeaterTempMax, "hx", "htMax", "heaterTempMax"); err != nil {
		return err
	}
	if err := decode(&settings.CO2WarnPPM, "cw", "co2WarnPPM"); err != nil {
		return err
	}
	if err := decode(&settings.NH3WarnPPM, "nw", "nh3WarnPPM"); err != nil {
		return err
	}
	if err := decode(&settings.CO2MinPct, "c2min", "co2MinPct"); err != nil {
		return err
	}
	if err := decode(&settings.CO2MaxPct, "c2max", "co2MaxPct"); err != nil {
		return err
	}
	if err := decode(&settings.NH3MinPct, "n3min", "nh3MinPct"); err != nil {
		return err
	}
	if err := decode(&settings.NH3MaxPct, "n3max", "nh3MaxPct"); err != nil {
		return err
	}
	if err := decode(&settings.DurationHours, "du", "dur", "durationHours"); err != nil {
		return err
	}
	if err := decode(&settings.FanOnDuration, "on", "fon", "fanOnDuration"); err != nil {
		return err
	}
	if err := decode(&settings.FanOffDuration, "of", "foff", "fanOffDuration"); err != nil {
		return err
	}
	return decode(&settings.DehumidifierHum, "dh", "dehum", "dehumidifierHum")
}

func preferredField(fields map[string]json.RawMessage, keys ...string) (json.RawMessage, bool) {
	for _, key := range keys {
		if raw, ok := fields[key]; ok && string(raw) != "null" {
			return raw, true
		}
	}
	return nil, false
}

func preferredKey(fields map[string]json.RawMessage, keys ...string) bool {
	for _, key := range keys {
		if _, ok := fields[key]; ok {
			return true
		}
	}
	return false
}
