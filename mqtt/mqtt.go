package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

		log.Println("✅ MQTT Connected")

		SubscribeTelemetry()
		SubscribeNotification()
	}

	// ========================================================
	// CONNECTION LOST
	// ========================================================

	opts.OnConnectionLost = func(client mqtt.Client, err error) {

		utils.ErrorLog.Println("❌ MQTT Connection Lost:", err)
	}

	// ========================================================
	// RECONNECT
	// ========================================================

	opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {

		log.Println("🔄 MQTT Reconnecting...")
	}

	// ========================================================
	// CREATE CLIENT
	// ========================================================

	service.MqttClient = mqtt.NewClient(opts)

	// ========================================================
	// CONNECT
	// ========================================================

	log.Println("🔌 Connecting MQTT Broker...")

	token := service.MqttClient.Connect()

	if token.WaitTimeout(15 * time.Second) {

		if token.Error() != nil {

			utils.ErrorLog.Println("❌ MQTT Connect Error:", token.Error())
			return
		}

	} else {

		utils.ErrorLog.Println("❌ MQTT Connection Timeout")
		return
	}
}

// ============================================================
// SUBSCRIBE TELEMETRY
// ============================================================

func SubscribeTelemetry() {

	if service.MqttClient == nil {

		utils.ErrorLog.Println("❌ MQTT Client Nil")
		return
	}

	if !service.MqttClient.IsConnected() {

		utils.ErrorLog.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/telemetry/+"

	log.Println("📡 Subscribing:", topic)

	token := service.MqttClient.Subscribe(
		topic,
		0,
		MessageHandler,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {

			utils.ErrorLog.Println("❌ Subscribe Error:", token.Error())
			return
		}

		log.Println("✅ Subscribed:", topic)

	} else {

		utils.ErrorLog.Println("❌ Subscribe Timeout")
	}
}

func SubscribeNotification() {

	if service.MqttClient == nil {

		utils.ErrorLog.Println("❌ MQTT Client Nil")
		return
	}

	if !service.MqttClient.IsConnected() {

		utils.ErrorLog.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/notification/+"

	log.Println("📡 Subscribing:", topic)

	token := service.MqttClient.Subscribe(
		topic,
		0,
		NotificationHandler,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {

			utils.ErrorLog.Println("❌ Subscribe Error:", token.Error())
			return
		}

		log.Println("✅ Subscribed:", topic)

	} else {

		utils.ErrorLog.Println("❌ Subscribe Timeout")
	}
}

// ============================================================
// MQTT MESSAGE HANDLER
// ============================================================

func MessageHandler(client mqtt.Client, msg mqtt.Message) {

	log.Println("\n📨 MQTT Message")
	log.Println("Topic:", msg.Topic())
	log.Println("Payload:", string(msg.Payload()))

	var telemetry model.Telemetry

	err := json.Unmarshal(msg.Payload(), &telemetry)

	if err != nil {
		utils.ErrorLog.Println("❌ JSON Error:", err)
		return
	}

	if telemetry.DeviceID == "" {

		utils.ErrorLog.Println("❌ Device ID Missing")
		return
	}

	go service.UpsetTelemetry(context.Background(), telemetry)

	log.Println("✅ Telemetry Updated For:", telemetry.DeviceID)

}

func NotificationHandler(client mqtt.Client, msg mqtt.Message) {

	log.Println("\n📨 MQTT Notification")
	log.Println("Topic:", msg.Topic())
	log.Println("Payload:", string(msg.Payload()))

	var notification model.Notification

	err := json.Unmarshal(msg.Payload(), &notification)

	if err != nil {
		utils.ErrorLog.Println("❌ JSON Error:", err)
		return
	}

	if notification.DeviceID == "" {
		utils.ErrorLog.Println("❌ Device ID Missing")
		return
	}

	go sendNotification(notification)

	log.Println("✅ Telemetry Updated For:", notification)
}

// ============================================================
// SEND MQTT COMMAND
// ============================================================

func SendCommand(deviceID string, payload interface{}) {

	if service.MqttClient == nil {

		utils.ErrorLog.Println("❌ MQTT Client Nil")
		return
	}

	if !service.MqttClient.IsConnected() {

		utils.ErrorLog.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/cmd/" + deviceID

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		utils.ErrorLog.Println("❌ JSON Marshal Error:", err)
		return
	}

	log.Println("📤 Publishing To:", topic)
	log.Println("Payload:", string(jsonPayload))

	token := service.MqttClient.Publish(
		topic,
		0,
		false,
		jsonPayload,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {
			utils.ErrorLog.Println("❌ Publish Error:", token.Error())
			return
		}

		log.Println("✅ Command Sent")

	} else {

		utils.ErrorLog.Println("❌ Publish Timeout")
	}
}

func sendNotification(notification model.Notification) {
	log.Printf("🔔 New Notification: Type=%s, DeviceID=%s, Timestamp=%d\n",
		notification.Type, notification.DeviceID, notification.Timestamp)

	go saveNotification(notification)

	user, err := service.GetUserByDeviceID(context.Background(), notification.DeviceID)
	if err != nil {
		utils.ErrorLog.Println("❌ Error fetching user:", err)
		return
	}

	log.Printf("🔔 Notification for User: %s (%s)\n", user.Username, user.DeviceID)

	if user.FcmToken == "" {
		utils.ErrorLog.Println("❌ User has no FCM token:", user.Username)
		return
	}

	err = SendFirebaseNotification(user.FcmToken, notification)
	if err != nil {
		utils.ErrorLog.Println("❌ Error sending Firebase notification:", err)
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
		return err
	}

	fmt.Println("Successfully sent:", response)

	return nil
}

func saveNotification(notification model.Notification) {
	notification.CreatedAt = time.Now()

	_, err := database.Notification.InsertOne(context.Background(), notification)
	if err != nil {
		utils.ErrorLog.Println("❌ Error saving notification:", err)
		return
	}
}
