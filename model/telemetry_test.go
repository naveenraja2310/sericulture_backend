package model

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTelemetryUnmarshalSupportsOldAndCompactPayloads(t *testing.T) {
	oldPayload := []byte(`{
		"temperature": 28.5, "humidity": 70, "motor": 1, "fan": 0,
		"heater": 1, "deviceId": "ESP001", "sensorFailure": false,
		"productionCompleted": false, "systemEnabled": false,
		"stages": [{"durationHours": 4, "fanTempMin": 20, "fanOnDuration": 0}]
	}`)
	newPayload := []byte(`{
		"t": 28.5, "h": 70, "m": 1, "f": 0, "ht": 1, "id": "ESP001",
		"senFail": 0, "prodDone": 0, "en": 0, "relaysOn": 0,
		"stg": 2, "dur": 0, "elap": 0, "rem": 0,
		"dehum": 0, "dehAct": 0,
		"s": [{"dur": 0, "ftMin": 0, "fon": 0, "foff": 0}]
	}`)

	var oldTelemetry, newTelemetry Telemetry
	if err := json.Unmarshal(oldPayload, &oldTelemetry); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(newPayload, &newTelemetry); err != nil {
		t.Fatal(err)
	}

	if oldTelemetry.DeviceID != "ESP001" || newTelemetry.DeviceID != "ESP001" {
		t.Fatalf("device IDs = %q and %q", oldTelemetry.DeviceID, newTelemetry.DeviceID)
	}
	if newTelemetry.Motor != 1 || newTelemetry.Fan != 0 || newTelemetry.SensorFailure {
		t.Fatalf("compact readings decoded incorrectly: %+v", newTelemetry)
	}
	if newTelemetry.PowerOn != 0 || newTelemetry.RelaysOn != 0 || newTelemetry.DehumidifierActive {
		t.Fatalf("compact zero values decoded incorrectly: %+v", newTelemetry)
	}
	if newTelemetry.StageDurationHours != 0 || len(newTelemetry.Stages) != 1 || newTelemetry.Stages[0].DurationHours != 0 || newTelemetry.Stages[0].DehumidifierHum != 0 {
		t.Fatalf("compact stage values decoded incorrectly: %+v", newTelemetry)
	}
	if oldTelemetry.Stages[0].DurationHours != 4 || oldTelemetry.Stages[0].FanTempMin != 20 {
		t.Fatalf("old stage values decoded incorrectly: %+v", oldTelemetry.Stages[0])
	}
}

func TestTelemetryUnmarshalPrefersCompactKeysAndProtectsMongoID(t *testing.T) {
	payload := []byte(`{
		"id": "device-from-short-key", "deviceId": "device-from-old-key",
		"temperature": 10, "t": 20,
		"powerOn": 1, "relaysOn": 0,
		"stages": [{"durationHours": 5, "dur": 0}]
	}`)

	var telemetry Telemetry
	if err := json.Unmarshal(payload, &telemetry); err != nil {
		t.Fatal(err)
	}

	if telemetry.ID != (primitive.ObjectID{}) {
		t.Fatalf("JSON device id populated MongoDB ID: %v", telemetry.ID)
	}
	if telemetry.DeviceID != "device-from-short-key" || telemetry.Temperature != 20 {
		t.Fatalf("compact keys did not win: %+v", telemetry)
	}
	if telemetry.PowerOn != 0 || telemetry.Stages[0].DurationHours != 0 {
		t.Fatalf("compact zero values did not win: %+v", telemetry)
	}
}

func TestTelemetryUnmarshalNormalizesMode(t *testing.T) {
	for input, expected := range map[string]string{
		"A": "AUTO", "auto": "AUTO", "M": "MANUAL", "manual": "MANUAL",
	} {
		var telemetry Telemetry
		if err := json.Unmarshal([]byte(`{"mode":"`+input+`"}`), &telemetry); err != nil {
			t.Fatal(err)
		}
		if telemetry.Mode != expected {
			t.Errorf("mode %q decoded as %q, want %q", input, telemetry.Mode, expected)
		}
	}
}

func TestTelemetryUnmarshalSupportsOldStatusFields(t *testing.T) {
	var telemetry Telemetry
	err := json.Unmarshal([]byte(`{
		"mode":"AUTO", "gprsStatus":"CONNECTED", "stageNumber":3,
		"systemStatus":"ENABLED", "relaysActive":1
	}`), &telemetry)
	if err != nil {
		t.Fatal(err)
	}

	if telemetry.Mode != "AUTO" || telemetry.GprsStatus != "CONNECTED" ||
		telemetry.ActiveStage != 3 || telemetry.StageNumber != 3 ||
		telemetry.SystemStatus != "ENABLED" || telemetry.RelaysActive != 1 {
		t.Fatalf("old status fields decoded incorrectly: %+v", telemetry)
	}
}

func TestTelemetryUnmarshalNormalizesCompactStatuses(t *testing.T) {
	var telemetry Telemetry
	err := json.Unmarshal([]byte(`{"mode":"A", "gprs":1, "en":0, "relaysOn":0}`), &telemetry)
	if err != nil {
		t.Fatal(err)
	}

	if telemetry.Mode != "AUTO" || telemetry.GprsStatus != "CONNECTED" ||
		telemetry.SystemStatus != "DISABLED" || telemetry.SystemEnabled {
		t.Fatalf("compact statuses decoded incorrectly: %+v", telemetry)
	}
}

func TestTelemetryUnmarshalSupportsAirQualityPayload(t *testing.T) {
	payload := []byte(`{
		"co2": 420.5, "co2Raw": 400, "co2Pct": 0.04, "co2st": 1, "co2Warn": 1,
		"nh3": 12.5, "nh3Raw": 10, "nh3Pct": 0.02, "nh3Rs": 150, "nh3Ratio": 0.8, "nh3Warn": 0,
		"tOff": 1.5, "hOff": -2, "co2Off": 3, "nh3Off": 4,
		"co2Set": 500, "nh3Set": 30, "co2Min": 0.01, "co2Max": 0.05,
		"nh3Min": 0.02, "nh3Max": 0.06, "co2Ovr": 1, "nh3Ovr": 0, "airOvr": 1,
		"s": [{
			"cw": 510, "nw": 31, "c2min": 0.011, "c2max": 0.051,
			"n3min": 0.021, "n3max": 0.061
		}]
	}`)

	var telemetry Telemetry
	if err := json.Unmarshal(payload, &telemetry); err != nil {
		t.Fatal(err)
	}

	if telemetry.CO2 != 420.5 || telemetry.CO2Raw != 400 || telemetry.CO2Pct != 0.04 ||
		telemetry.CO2Status != "1" || !telemetry.CO2Warn {
		t.Errorf("CO2 values decoded incorrectly: %+v", telemetry)
	}
	if telemetry.NH3 != 12.5 || telemetry.NH3Raw != 10 || telemetry.NH3Pct != 0.02 ||
		telemetry.NH3Rs != 150 || telemetry.NH3Ratio != 0.8 || telemetry.NH3Warn {
		t.Errorf("NH3 values decoded incorrectly: %+v", telemetry)
	}
	if telemetry.TempOffset != 1.5 || telemetry.HumOffset != -2 ||
		telemetry.CO2Offset != 3 || telemetry.NH3Offset != 4 {
		t.Errorf("sensor offsets decoded incorrectly: %+v", telemetry)
	}
	if telemetry.CO2WarnPPM != 500 || telemetry.NH3WarnPPM != 30 ||
		telemetry.CO2MinPct != 0.01 || telemetry.CO2MaxPct != 0.05 ||
		telemetry.NH3MinPct != 0.02 || telemetry.NH3MaxPct != 0.06 {
		t.Errorf("air quality thresholds decoded incorrectly: %+v", telemetry)
	}
	if !telemetry.CO2FanOverride || telemetry.NH3FanOverride || !telemetry.AirOverride {
		t.Errorf("air overrides decoded incorrectly: %+v", telemetry)
	}
	if len(telemetry.Stages) != 1 {
		t.Fatalf("got %d stages, want 1", len(telemetry.Stages))
	}
	stage := telemetry.Stages[0]
	if stage.CO2WarnPPM != 510 || stage.NH3WarnPPM != 31 ||
		stage.CO2MinPct != 0.011 || stage.CO2MaxPct != 0.051 ||
		stage.NH3MinPct != 0.021 || stage.NH3MaxPct != 0.061 {
		t.Errorf("stage air quality settings decoded incorrectly: %+v", stage)
	}
}
