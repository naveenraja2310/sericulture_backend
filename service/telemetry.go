package service

import (
	"context"
	"log"
	"sericulture/database"
	"sericulture/model"
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

		"gprsStatus": telemetry.GprsStatus,
		"deviceId":   telemetry.DeviceID,
		"uptime":     telemetry.Uptime,
		"powerOn":    telemetry.PowerOn,

		"mode":          telemetry.Mode,
		"tempThreshold": telemetry.TempThreshold,
		"humThreshold":  telemetry.HumThreshold,
		"fanCycleTime":  telemetry.FanCycleTime,

		"activeStage":         telemetry.ActiveStage,
		"stageTemp":           telemetry.StageTemp,
		"stageHum":            telemetry.StageHum,
		"stageElapsedHours":   telemetry.StageElapsedHours,
		"stageRemainingHours": telemetry.StageRemainingHours,
		"stages":              telemetry.Stages,

		"lastTelemetryTimestamp": telemetry.LastTelemetryTimestamp,
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
			log.Println("❌ Removed slow websocket client")
		}
	}
}

func GetTelemetry(ctx context.Context, deviceID string) (*model.Telemetry, error) {
	var telemetry model.Telemetry
	err := database.Telemetry.FindOne(ctx, bson.M{"deviceId": deviceID}).Decode(&telemetry)
	if err != nil {
		return nil, err
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
