package service

import (
	"context"
	"sericulture/database"
	"sericulture/model"
	"sericulture/utils"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WsClient struct {
	Conn *websocket.Conn
	Send chan interface{}
}

var (
	// deviceID -> clients
	DeviceClients = map[string]map[*WsClient]bool{}
	Mutex         sync.RWMutex
	MqttClient    mqtt.Client
)

func UpsetTelemetry(ctx context.Context, telemetry model.Telemetry) (*model.Telemetry, error) {

	updateFields := bson.M{
		"temperature": telemetry.Temperature,
		"humidity":    telemetry.Humidity,
		"motor":       telemetry.Motor,
		"fan":         telemetry.Fan,
		"heater":      telemetry.Heater,
		"spare":       telemetry.Spare,

		"gprsStatus":         telemetry.GprsStatus,
		"deviceId":           telemetry.DeviceID,
		"uptime":             telemetry.Uptime,
		"powerOn":            telemetry.PowerOn,
		"sensorFailure":      telemetry.SensorFailure,
		"timer":              telemetry.Timer,
		"dehumidifierHum":    telemetry.DehumidifierHum,
		"dehumidifierActive": telemetry.DehumidifierActive,
		"relaysOn":           telemetry.RelaysOn,
		"relaysActive":       telemetry.RelaysActive,
		"systemStatus":       telemetry.SystemStatus,

		"mode": telemetry.Mode,

		"fanTempMin":     telemetry.FanTempMin,
		"fanTempMax":     telemetry.FanTempMax,
		"motorHumMin":    telemetry.MotorHumMin,
		"motorHumMax":    telemetry.MotorHumMax,
		"heaterTempMin":  telemetry.HeaterTempMin,
		"heaterTempMax":  telemetry.HeaterTempMax,
		"fanOnDuration":  telemetry.FanOnDuration,
		"fanOffDuration": telemetry.FanOffDuration,

		"activeStage":         telemetry.ActiveStage,
		"stageNumber":         telemetry.StageNumber,
		"stageDurationHours":  telemetry.StageDurationHours,
		"stageElapsedHours":   telemetry.StageElapsedHours,
		"stageRemainingHours": telemetry.StageRemainingHours,
		"stages":              telemetry.Stages,

		"lastTelemetryTimestamp": telemetry.LastTelemetryTimestamp,
		"productionCompleted":    telemetry.ProductionCompleted,
		"systemEnabled":          telemetry.SystemEnabled,
		"firmwareVersion":        telemetry.FirmwareVersion,
		"updatedAt":              time.Now(),
	}

	update := bson.M{
		"$set": updateFields,
	}

	_, err := database.Telemetry.UpdateOne(ctx, bson.M{"deviceId": telemetry.DeviceID}, update, options.Update().SetUpsert(true))

	if err != nil {
		return nil, err
	}

	// Broadcast to websocket clients
	BroadcastTelemetry(telemetry.DeviceID, telemetry)

	return &telemetry, nil
}

// Broadcast telemetry to all websocket clients of a device
func BroadcastTelemetry(deviceID string, telemetry interface{}) {

	// Copy current clients under read lock to avoid concurrent map access
	Mutex.RLock()
	clientsMap := DeviceClients[deviceID]
	if clientsMap == nil {
		Mutex.RUnlock()
		return
	}

	clients := make([]*WsClient, 0, len(clientsMap))
	for c := range clientsMap {
		clients = append(clients, c)
	}
	Mutex.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- telemetry:
		default:
			// Slow/dead client cleanup: take write lock to modify shared map
			Mutex.Lock()

			// re-check map exists and client still present
			if m, ok := DeviceClients[deviceID]; ok {
				delete(m, client)
				if len(m) == 0 {
					delete(DeviceClients, deviceID)
				}
			}

			Mutex.Unlock()
			close(client.Send)
			_ = client.Conn.Close()
			utils.Warn(nil, "Removed slow websocket client", "deviceId", deviceID)
		}
	}
}

func GetTelemetry(ctx context.Context, deviceID string) (*model.Telemetry, error) {
	var telemetry model.Telemetry
	err := database.Telemetry.FindOne(ctx, bson.M{"deviceId": deviceID}).Decode(&telemetry)
	if err != nil {
		return nil, err
	}
	telemetry.Mode = model.NormalizeMode(telemetry.Mode)
	telemetry.GprsStatus = model.NormalizeGprsStatus(telemetry.GprsStatus)
	telemetry.SystemStatus = model.NormalizeSystemStatus(telemetry.SystemStatus)
	if telemetry.SystemStatus == "" {
		if telemetry.SystemEnabled {
			telemetry.SystemStatus = "ENABLED"
		} else {
			telemetry.SystemStatus = "DISABLED"
		}
	}
	return &telemetry, nil
}

func DeleteTelemetry(ctx context.Context, deviceID string) (*model.Telemetry, error) {
	_, err := database.Telemetry.DeleteOne(ctx, bson.M{"deviceId": deviceID})
	if err != nil {
		return nil, err
	}
	return &model.Telemetry{}, nil
}

func GetAllTelemetry(ctx context.Context, limit, offset int64, search string) ([]map[string]interface{}, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"deviceId": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Get total count
	total, err := database.Telemetry.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Aggregation pipeline with lookup to join user collection
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "deviceId",
				"foreignField": "deviceId",
				"as":           "userDetails",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$userDetails",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$addFields": bson.M{
				"username": "$userDetails.username",
			},
		},
		{"$skip": offset},
		{"$limit": limit},
	}

	cursor, err := database.Telemetry.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var telemetries []map[string]interface{}
	if err = cursor.All(ctx, &telemetries); err != nil {
		return nil, 0, err
	}

	return telemetries, total, nil
}
