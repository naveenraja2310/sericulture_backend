package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	websocket "github.com/gofiber/websocket/v2"
)

// ============================================================
// TELEMETRY MODEL
// ============================================================

type Telemetry struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`

	Motor  int `json:"motor"`
	Fan    int `json:"fan"`
	Heater int `json:"heater"`
	Spare  int `json:"spare"`

	Mode string `json:"mode"`

	TempThreshold float64 `json:"tempThreshold"`
	HumThreshold  float64 `json:"humThreshold"`

	FanCycleTime int `json:"fanCycleTime"`

	Uptime int `json:"uptime"`

	GprsStatus string `json:"gprsStatus"`

	DeviceID string `json:"deviceId"`

	PowerOn int `json:"powerOn"`
}

// ============================================================
// GLOBALS
// ============================================================

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

var (
	mqttClient mqtt.Client

	mutex sync.Mutex

	deviceTelemetry = map[string]Telemetry{}

	subscriptionReady = make(chan bool, 1)
	deviceClients     = map[string]map[*websocket.Conn]*wsClient{}
)

// ============================================================
// MAIN
// ============================================================

func main() {

	connectMQTT()

	<-subscriptionReady

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	app.Use(logger.New())
	app.Use(recover.New())
	// ========================================================
	// HEALTH
	// ========================================================

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// ========================================================
	// GET DEVICE STATUS (HTTP)
	// ========================================================

	app.Get("/device/:id/status", getDeviceStatus)

	// ========================================================
	// GET DEVICE STATUS (WebSocket)
	// Clients can connect to `/device/:id/ws` to receive live updates
	// ========================================================

	app.Get("/device/:id/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("deviceID", c.Params("id"))
			return c.Next()
		}
		return c.Status(fiber.StatusUpgradeRequired).SendString("Upgrade Required")
	}, websocket.New(deviceStatusWS))

	// ========================================================
	// MANUAL DEVICE CONTROL
	// ========================================================

	app.Post("/device/:id/fan/on", fanOn)
	app.Post("/device/:id/fan/off", fanOff)

	app.Post("/device/:id/motor/on", motorOn)
	app.Post("/device/:id/motor/off", motorOff)

	app.Post("/device/:id/heater/on", heaterOn)
	app.Post("/device/:id/heater/off", heaterOff)

	// ========================================================
	// SYSTEM MODE
	// ========================================================

	app.Post("/device/:id/mode/auto", setAutoMode)
	app.Post("/device/:id/mode/manual", setManualMode)

	// ========================================================
	// THRESHOLDS
	// ========================================================

	app.Post("/device/:id/temp-threshold", setTempThreshold)

	app.Post("/device/:id/hum-threshold", setHumThreshold)

	// ========================================================
	// FAN TIMER
	// ========================================================

	app.Post("/device/:id/fan-cycle", setFanCycle)

	// ========================================================
	// START SERVER
	// ========================================================

	fmt.Println("🚀 HTTP Server Started On Port 3000")

	err := app.Listen(":3000")

	if err != nil {
		fmt.Println("❌ Fiber Error:", err)
	}
}

// ============================================================
// MQTT CONNECT
// ============================================================

func connectMQTT() {

	opts := mqtt.NewClientOptions()

	// ========================================================
	// BROKER
	// ========================================================

	opts.AddBroker("tcp://13.205.144.42:1883")

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

		fmt.Println("✅ MQTT Connected")

		go subscribeTelemetry()
	}

	// ========================================================
	// CONNECTION LOST
	// ========================================================

	opts.OnConnectionLost = func(client mqtt.Client, err error) {

		fmt.Println("❌ MQTT Connection Lost:", err)
	}

	// ========================================================
	// RECONNECT
	// ========================================================

	opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {

		fmt.Println("🔄 MQTT Reconnecting...")
	}

	// ========================================================
	// CREATE CLIENT
	// ========================================================

	mqttClient = mqtt.NewClient(opts)

	// ========================================================
	// CONNECT
	// ========================================================

	fmt.Println("🔌 Connecting MQTT Broker...")

	token := mqttClient.Connect()

	if token.WaitTimeout(15 * time.Second) {

		if token.Error() != nil {

			fmt.Println("❌ MQTT Connect Error:", token.Error())
			return
		}

	} else {

		fmt.Println("❌ MQTT Connection Timeout")
		return
	}
}

// ============================================================
// SUBSCRIBE TELEMETRY
// ============================================================

func subscribeTelemetry() {

	if mqttClient == nil {

		fmt.Println("❌ MQTT Client Nil")
		return
	}

	if !mqttClient.IsConnected() {

		fmt.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/telemetry/+"

	fmt.Println("📡 Subscribing:", topic)

	token := mqttClient.Subscribe(
		topic,
		0,
		messageHandler,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {

			fmt.Println("❌ Subscribe Error:", token.Error())
			return
		}

		fmt.Println("✅ Subscribed:", topic)

		subscriptionReady <- true

	} else {

		fmt.Println("❌ Subscribe Timeout")
	}
}

// ============================================================
// MQTT MESSAGE HANDLER
// ============================================================

func messageHandler(client mqtt.Client, msg mqtt.Message) {

	fmt.Println("\n📨 MQTT Message")
	fmt.Println("Topic:", msg.Topic())
	fmt.Println("Payload:", string(msg.Payload()))

	var telemetry Telemetry

	err := json.Unmarshal(msg.Payload(), &telemetry)

	if err != nil {

		fmt.Println("❌ JSON Error:", err)
		return
	}

	if telemetry.DeviceID == "" {

		fmt.Println("❌ Device ID Missing")
		return
	}

	mutex.Lock()

	deviceTelemetry[telemetry.DeviceID] = telemetry

	mutex.Unlock()

	fmt.Println("✅ Telemetry Updated For:", telemetry.DeviceID)

	// Broadcast telemetry update to any connected websocket clients
	mutex.Lock()
	conns := deviceClients[telemetry.DeviceID]
	var clients []*wsClient
	for _, client := range conns {
		clients = append(clients, client)
	}
	mutex.Unlock()

	for _, client := range clients {
		client.mu.Lock()
		if err := client.conn.WriteJSON(telemetry); err != nil {
			fmt.Println("❌ WebSocket Write Error:", err)
		}
		client.mu.Unlock()
	}
}

// ============================================================
// SEND MQTT COMMAND
// ============================================================

func sendCommand(deviceID string, payload interface{}) {

	if mqttClient == nil {

		fmt.Println("❌ MQTT Client Nil")
		return
	}

	if !mqttClient.IsConnected() {

		fmt.Println("❌ MQTT Not Connected")
		return
	}

	topic := "/cmd/" + deviceID

	jsonPayload, err := json.Marshal(payload)

	if err != nil {

		fmt.Println("❌ JSON Marshal Error:", err)
		return
	}

	fmt.Println("📤 Publishing To:", topic)
	fmt.Println("Payload:", string(jsonPayload))

	token := mqttClient.Publish(
		topic,
		0,
		false,
		jsonPayload,
	)

	if token.WaitTimeout(10 * time.Second) {

		if token.Error() != nil {

			fmt.Println("❌ Publish Error:", token.Error())
			return
		}

		fmt.Println("✅ Command Sent")

	} else {

		fmt.Println("❌ Publish Timeout")
	}
}

// ============================================================
// GET DEVICE STATUS
// ============================================================

func getDeviceStatus(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	// delete cached telemetry to force device to send fresh status via MQTT
	mutex.Lock()
	delete(deviceTelemetry, deviceID)
	mutex.Unlock()

	payload := fiber.Map{
		"method": "getStatus",
	}

	sendCommand(deviceID, payload)

	return c.JSON(Telemetry{})
}

// ============================================================
// WEBSOCKET: DEVICE STATUS
// ============================================================

func deviceStatusWS(c *websocket.Conn) {
	// Get device ID from fiber.Ctx locals set before upgrade
	v := c.Locals("deviceID")
	deviceID, _ := v.(string)
	if deviceID == "" {
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "device id missing"))
		return
	}

	client := &wsClient{conn: c}

	// Register client
	mutex.Lock()
	if deviceClients[deviceID] == nil {
		deviceClients[deviceID] = make(map[*websocket.Conn]*wsClient)
	}
	deviceClients[deviceID][c] = client
	data, ok := deviceTelemetry[deviceID]
	mutex.Unlock()

	// Send initial status if available
	if ok {
		client.mu.Lock()
		_ = client.conn.WriteJSON(data)
		client.mu.Unlock()
	}

	// Keep connection open; read messages so TCP stays alive
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}

	// Unregister client
	mutex.Lock()
	delete(deviceClients[deviceID], c)
	mutex.Unlock()
	c.Close()
}

// ============================================================
// DEVICE CONTROL
// ============================================================

func sendDeviceControl(deviceID string, device string, state int) {

	payload := fiber.Map{
		"device": device,
		"state":  state,
	}

	sendCommand(deviceID, payload)
}

// ============================================================
// FAN
// ============================================================

func fanOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "fan", 1)

	return c.JSON(fiber.Map{
		"message": "fan on",
	})
}

func fanOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "fan", 0)

	return c.JSON(fiber.Map{
		"message": "fan off",
	})
}

// ============================================================
// MOTOR
// ============================================================

func motorOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "motor", 1)

	return c.JSON(fiber.Map{
		"message": "motor on",
	})
}

func motorOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "motor", 0)

	return c.JSON(fiber.Map{
		"message": "motor off",
	})
}

// ============================================================
// HEATER
// ============================================================

func heaterOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "heater", 1)

	return c.JSON(fiber.Map{
		"message": "heater on",
	})
}

func heaterOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	sendDeviceControl(deviceID, "heater", 0)

	return c.JSON(fiber.Map{
		"message": "heater off",
	})
}

// ============================================================
// MODE CONTROL
// ============================================================

func setAutoMode(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	payload := fiber.Map{
		"method": "setMode",
		"params": true,
	}

	sendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "AUTO mode enabled",
	})
}

func setManualMode(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	payload := fiber.Map{
		"method": "setMode",
		"params": false,
	}

	sendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "MANUAL mode enabled",
	})
}

// ============================================================
// TEMP THRESHOLD
// ============================================================

type TempThresholdRequest struct {
	Value float64 `json:"value"`
}

func setTempThreshold(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	var body TempThresholdRequest

	if err := c.BodyParser(&body); err != nil {

		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	payload := fiber.Map{
		"method": "setTempThreshold",
		"params": body.Value,
	}

	sendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "temperature threshold updated",
	})
}

// ============================================================
// HUM THRESHOLD
// ============================================================

type HumThresholdRequest struct {
	Value float64 `json:"value"`
}

func setHumThreshold(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	var body HumThresholdRequest

	if err := c.BodyParser(&body); err != nil {

		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	payload := fiber.Map{
		"method": "setHumThreshold",
		"params": body.Value,
	}

	sendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "humidity threshold updated",
	})
}

// ============================================================
// FAN CYCLE
// ============================================================

type FanCycleRequest struct {
	Minutes int `json:"minutes"`
}

func setFanCycle(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	var body FanCycleRequest

	if err := c.BodyParser(&body); err != nil {

		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	payload := fiber.Map{
		"method": "setFanCycle",
		"params": body.Minutes,
	}

	sendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "fan cycle updated",
	})
}
