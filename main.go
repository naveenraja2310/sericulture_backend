package main

import (
	"encoding/json"
	"fmt"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
}

// ============================================================
// GLOBALS
// ============================================================

var (
	mqttClient mqtt.Client

	mutex sync.Mutex

	deviceTelemetry = map[string]Telemetry{}
)

// ============================================================
// MAIN
// ============================================================

func main() {

	connectMQTT()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// ========================================================
	// HEALTH
	// ========================================================

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// ========================================================
	// GET DEVICE STATUS
	// ========================================================

	app.Get("/device/:id/status", getDeviceStatus)

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

	app.Listen(":3000")
}

// ============================================================
// MQTT CONNECT
// ============================================================

func connectMQTT() {

	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://13.205.144.42:1883")

	opts.SetClientID("golang-backend")

	opts.AutoReconnect = true

	opts.OnConnect = func(client mqtt.Client) {

		fmt.Println("✅ MQTT Connected")

		subscribeTelemetry()
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {

		fmt.Println("❌ MQTT Connection Lost:", err)
	}

	mqttClient = mqtt.NewClient(opts)

	token := mqttClient.Connect()

	token.Wait()

	if token.Error() != nil {
		panic(token.Error())
	}
}

// ============================================================
// SUBSCRIBE
// ============================================================

func subscribeTelemetry() {

	topic := "/telemetry/+"

	token := mqttClient.Subscribe(
		topic,
		0,
		messageHandler,
	)

	token.Wait()

	if token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("📡 Subscribed:", topic)
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

	mutex.Lock()

	deviceTelemetry[telemetry.DeviceID] = telemetry

	mutex.Unlock()

	fmt.Println("✅ Telemetry Updated For:", telemetry.DeviceID)
}

// ============================================================
// SEND MQTT COMMAND
// ============================================================

func sendCommand(deviceID string, payload interface{}) {

	topic := "/cmd/" + deviceID

	jsonPayload, _ := json.Marshal(payload)

	token := mqttClient.Publish(
		topic,
		0,
		false,
		jsonPayload,
	)

	token.Wait()

	if token.Error() != nil {

		fmt.Println("❌ Publish Error:", token.Error())

		return
	}

	fmt.Println("📤 Command Sent")
	fmt.Println(string(jsonPayload))
}

// ============================================================
// GET DEVICE STATUS
// ============================================================

func getDeviceStatus(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	mutex.Lock()

	defer mutex.Unlock()

	data, ok := deviceTelemetry[deviceID]

	if !ok {

		return c.Status(404).JSON(fiber.Map{
			"error": "device not found",
		})
	}

	return c.JSON(data)
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

		return err
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

		return err
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

		return err
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
