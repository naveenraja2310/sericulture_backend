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

	Mutex sync.RWMutex

	MqttClient mqtt.Client
)

func UpsetTelemetry(ctx context.Context, telemetry model.Telemetry) (*model.Telemetry, error) {

	updateFields := bson.M{
		"updatedAt":     time.Now(),
		"temperature":   telemetry.Temperature,
		"humidity":      telemetry.Humidity,
		"motor":         telemetry.Motor,
		"fan":           telemetry.Fan,
		"heater":        telemetry.Heater,
		"spare":         telemetry.Spare,
		"mode":          telemetry.Mode,
		"tempThreshold": telemetry.TempThreshold,
		"humThreshold":  telemetry.HumThreshold,
		"fanCycleTime":  telemetry.FanCycleTime,
		"uptime":        telemetry.Uptime,
		"gprsStatus":    telemetry.GprsStatus,
		"deviceId":      telemetry.DeviceID,
		"powerOn":       telemetry.PowerOn,
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

	Mutex.RLock()
	clients := DeviceClients[deviceID]
	Mutex.RUnlock()

	for client := range clients {
		select {
		case client.Send <- telemetry:
		default:
			// Slow/dead client cleanup
			Mutex.Lock()

			delete(DeviceClients[deviceID], client)

			if len(DeviceClients[deviceID]) == 0 {
				delete(DeviceClients, deviceID)
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
