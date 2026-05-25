package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sericulture/model"
	"sericulture/service"
	"time"

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
	}

	// ========================================================
	// CONNECTION LOST
	// ========================================================

	opts.OnConnectionLost = func(client mqtt.Client, err error) {

		log.Println("❌ MQTT Connection Lost:", err)
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

			log.Println("❌ MQTT Connect Error:", token.Error())
			return
		}

	} else {

		log.Println("❌ MQTT Connection Timeout")
		return
	}
}

// ============================================================
// SUBSCRIBE TELEMETRY
// ============================================================

func SubscribeTelemetry() {

	if service.MqttClient == nil {

		log.Println("❌ MQTT Client Nil")
		return
	}

	if !service.MqttClient.IsConnected() {

		log.Println("❌ MQTT Not Connected")
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

			log.Println("❌ Subscribe Error:", token.Error())
			return
		}

		log.Println("✅ Subscribed:", topic)

	} else {

		log.Println("❌ Subscribe Timeout")
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
		log.Println("❌ JSON Error:", err)
		return
	}

	if telemetry.DeviceID == "" {

		log.Println("❌ Device ID Missing")
		return
	}

	go service.UpsetTelemetry(context.Background(), telemetry)

	log.Println("✅ Telemetry Updated For:", telemetry.DeviceID)

}

// ============================================================
// SEND MQTT COMMAND
// ============================================================

func SendCommand(deviceID string, payload interface{}) {

	if service.MqttClient == nil {

		log.Println("❌ MQTT Client Nil")
		return
	}

	if !service.MqttClient.IsConnected() {

		log.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/cmd/" + deviceID

	jsonPayload, err := json.Marshal(payload)

	if err != nil {

		log.Println("❌ JSON Marshal Error:", err)
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

			log.Println("❌ Publish Error:", token.Error())
			return
		}

		log.Println("✅ Command Sent")

	} else {

		log.Println("❌ Publish Timeout")
	}
}
