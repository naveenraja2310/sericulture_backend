package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sericulture/database"
	"sericulture/firebase"
	"sericulture/model"
	"sericulture/service"
	"sericulture/utils"
	"time"

	"firebase.google.com/go/messaging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ============================================================
// MQTT CONNECT
// ============================================================

func ConnectMQTT(mqttUrl string) {

	opts := mqtt.NewClientOptions()

	// ========================================================
	// BROKER
	// ========================================================

	opts.AddBroker(mqttUrl)

	// ========================================================
	// UNIQUE CLIENT ID
	// ========================================================

	clientID := fmt.Sprintf(
		"golang-backend-%d",
		time.Now().Unix(),
	)

	opts.SetClientID(clientID)

	// ========================================================
	// MQTT SETTINGS
	// ========================================================

	opts.SetCleanSession(false)

	opts.AutoReconnect = true

	opts.ConnectRetry = true

	opts.ConnectRetryInterval = 5 * time.Second

	opts.KeepAlive = 60

	opts.PingTimeout = 10 * time.Second

	opts.Order = false

	// ========================================================
	// CONNECTED
	// ========================================================

	opts.OnConnect = func(client mqtt.Client) {

		utils.Info(nil, "MQTT connected")

		SubscribeTelemetry()
		SubscribeNotification()
	}

	// ========================================================
	// CONNECTION LOST
	// ========================================================

	opts.OnConnectionLost = func(client mqtt.Client, err error) {

		utils.Error(nil, "MQTT connection lost", "error", err.Error())
	}

	// ========================================================
	// RECONNECT
	// ========================================================

	opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {

		utils.Warn(nil, "MQTT reconnecting")
	}

	// ========================================================
	// CREATE CLIENT
	// ========================================================

	service.MqttClient = mqtt.NewClient(opts)

	// ========================================================
	// CONNECT
	// ========================================================

	utils.Info(nil, "Connecting MQTT broker")

	token := service.MqttClient.Connect()

	if token.WaitTimeout(15 * time.Second) {

		if token.Error() != nil {

			utils.Error(nil, "MQTT connect error", "error", token.Error().Error())
			return
		}

	} else {

		utils.Error(nil, "MQTT connection timeout")
		return
	}
}

// ============================================================
// SUBSCRIBE TELEMETRY
// ============================================================

func SubscribeTelemetry() {

	if service.MqttClient == nil {

		utils.Error(nil, "MQTT client nil while subscribing telemetry")
		return
	}

	if !service.MqttClient.IsConnected() {

		utils.Error(nil, "MQTT client not connected while subscribing telemetry")
		return
	}

	topic := "/telemetry/+"

	utils.Info(nil, "Subscribing to telemetry topic", "topic", topic)

	token := service.MqttClient.Subscribe(
		topic,
		0,
		MessageHandler,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {
			utils.Error(nil, "Telemetry subscription error", "topic", topic, "error", token.Error().Error())
			return
		}

		utils.Info(nil, "Telemetry subscription successful", "topic", topic)

	} else {
		utils.Error(nil, "Telemetry subscription timeout", "topic", topic)
	}
}

func SubscribeNotification() {

	if service.MqttClient == nil {

		utils.Error(nil, "MQTT client nil while subscribing notifications")
		return
	}

	if !service.MqttClient.IsConnected() {

		utils.Error(nil, "MQTT client not connected while subscribing notifications")
		return
	}

	topic := "/notification/+"

	utils.Info(nil, "Subscribing to notification topic", "topic", topic)

	token := service.MqttClient.Subscribe(
		topic,
		0,
		NotificationHandler,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {
			utils.Error(nil, "Notification subscription error", "topic", topic, "error", token.Error().Error())
			return
		}

		utils.Info(nil, "Notification subscription successful", "topic", topic)

	} else {
		utils.Error(nil, "Notification subscription timeout", "topic", topic)
	}
}

// ============================================================
// MQTT MESSAGE HANDLER
// ============================================================

func MessageHandler(client mqtt.Client, msg mqtt.Message) {

	utils.Info(nil, "MQTT message received", "topic", msg.Topic(), "payload", string(msg.Payload()))

	var telemetry model.Telemetry

	err := json.Unmarshal(msg.Payload(), &telemetry)

	if err != nil {
		utils.Error(nil, "Telemetry JSON parse error", "topic", msg.Topic(), "error", err.Error())
		return
	}

	if telemetry.DeviceID == "" {

		utils.Warn(nil, "Telemetry message missing device ID", "topic", msg.Topic())
		return
	}

	go service.UpsetTelemetry(context.Background(), telemetry)

	utils.Info(nil, "Telemetry updated", "deviceId", telemetry.DeviceID)

}

func NotificationHandler(client mqtt.Client, msg mqtt.Message) {

	utils.Info(nil, "MQTT notification received", "topic", msg.Topic(), "payload", string(msg.Payload()))

	var notification model.Notification

	err := json.Unmarshal(msg.Payload(), &notification)

	if err != nil {
		utils.Error(nil, "Notification JSON parse error", "topic", msg.Topic(), "error", err.Error())
		return
	}

	if notification.DeviceID == "" {
		utils.Warn(nil, "Notification message missing device ID", "topic", msg.Topic())
		return
	}

	go sendNotification(notification)

	utils.Info(nil, "Notification processed", "deviceId", notification.DeviceID)
}

// ============================================================
// SEND MQTT COMMAND
// ============================================================

func SendCommand(deviceID string, payload interface{}) {

	if service.MqttClient == nil {
		utils.Error(nil, "MQTT client nil while sending command", "deviceId", deviceID)
		return
	}

	if !service.MqttClient.IsConnected() {
		utils.Error(nil, "MQTT not connected while sending command", "deviceId", deviceID)
		return
	}

	topic := "/cmd/" + deviceID

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		utils.Error(nil, "Failed to marshal MQTT command payload", "deviceId", deviceID, "error", err.Error())
		return
	}

	utils.Info(nil, "Publishing MQTT command", "topic", topic, "payload", string(jsonPayload), "deviceId", deviceID)

	token := service.MqttClient.Publish(
		topic,
		0,
		false,
		jsonPayload,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {
			utils.Error(nil, "MQTT publish error", "topic", topic, "deviceId", deviceID, "error", token.Error().Error())
			return
		}

		utils.Info(nil, "MQTT command sent", "topic", topic, "deviceId", deviceID)

	} else {

		utils.Error(nil, "MQTT publish timeout", "topic", topic, "deviceId", deviceID)
	}
}

func sendNotification(notification model.Notification) {
	utils.Info(nil, "New notification received", "deviceId", notification.DeviceID, "type", notification.Type, "timestamp", notification.Timestamp)

	go saveNotification(notification)

	user, err := service.GetUserByDeviceID(context.Background(), notification.DeviceID)
	if err != nil {
		utils.Error(nil, "Failed to fetch user for notification", "deviceId", notification.DeviceID, "error", err.Error())
		return
	}

	utils.Info(nil, "Notification targeted for user", "userId", user.ID.Hex(), "deviceId", user.DeviceID, "username", user.Username)

	if user.FcmToken == "" {
		utils.Warn(nil, "User has no FCM token for notification", "userId", user.ID.Hex(), "deviceId", user.DeviceID)
		return
	}

	err = SendFirebaseNotification(user.FcmToken, notification)
	if err != nil {
		utils.Error(nil, "Failed to send Firebase notification", "userId", user.ID.Hex(), "deviceId", user.DeviceID, "error", err.Error())
		return
	}
}

func SendFirebaseNotification(token string, notification model.Notification) error {

	client, err := firebase.FirebaseClient.Messaging(context.Background())
	if err != nil {
		return err
	}

	// Send data-only message to avoid FCM/browser showing the notification
	// automatically and causing duplicates. The client (page or service
	// worker) will read these fields and display a single notification.
	// message := &messaging.Message{
	// 	Token: token,
	// 	Data: map[string]string{
	// 		"title": notification.Title,
	// 		"body":  notification.Body,
	// 		"url":   "/dashboard",
	// 	},
	// }

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
	}

	response, err := client.Send(context.Background(), message)
	if err != nil {
		utils.Error(nil, "Failed to send FCM notification", "token", token, "error", err.Error())
		return err
	}

	utils.Info(nil, "FCM notification sent successfully", "token", token, "response", response)

	return nil
}

func saveNotification(notification model.Notification) {
	notification.CreatedAt = time.Now()

	_, err := database.Notification.InsertOne(context.Background(), notification)
	if err != nil {
		utils.Error(nil, "Failed to save notification", "deviceId", notification.DeviceID, "error", err.Error())
		return
	}
}
