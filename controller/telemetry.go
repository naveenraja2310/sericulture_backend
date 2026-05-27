package controller

import (
	"context"
	"log"
	"net/http"
	"sericulture/database"
	"sericulture/model"
	"sericulture/mqtt"
	"sericulture/service"
	"sericulture/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
)

// ============================================================
// GET DEVICE STATUS
// ============================================================

func GetDeviceStatus(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	payload := fiber.Map{
		"method": "getStatus",
	}

	mqtt.SendCommand(deviceID, payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	telemetry, err := service.DeleteTelemetry(ctx, deviceID)
	if err != nil {
		return err
	}

	return c.JSON(telemetry)
}

// ============================================================
// WEBSOCKET: DEVICE STATUS
// ============================================================

func DeviceStatusWS(c *websocket.Conn) {
	// Get deviceID
	v := c.Locals("deviceID")
	deviceID, ok := v.(string)
	if !ok || deviceID == "" {
		_ = c.Close()
		return
	}

	client := &service.WsClient{
		Conn: c,
		Send: make(chan interface{}, 100),
	}

	// Register websocket client
	service.Mutex.Lock()

	if service.DeviceClients[deviceID] == nil {
		service.DeviceClients[deviceID] = make(map[*service.WsClient]bool)
	}

	service.DeviceClients[deviceID][client] = true

	service.Mutex.Unlock()

	log.Println("✅ WS Connected:", deviceID)

	// Send latest telemetry immediately
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	telemetry, err := service.GetTelemetry(ctx, deviceID)

	if err == nil {
		select {
		case client.Send <- telemetry:
		default:
		}
	}

	// Cleanup
	defer func() {
		service.Mutex.Lock()
		delete(service.DeviceClients[deviceID], client)
		if len(service.DeviceClients[deviceID]) == 0 {
			delete(service.DeviceClients, deviceID)
		}
		service.Mutex.Unlock()
		close(client.Send)
		_ = c.Close()
		utils.ErrorLog.Println("❌ WS Disconnected:", deviceID)
	}()

	// Read limits
	c.SetReadLimit(1024 * 10)

	_ = c.SetReadDeadline(
		time.Now().Add(pongWait),
	)

	c.SetPongHandler(func(string) error {
		return c.SetReadDeadline(
			time.Now().Add(pongWait),
		)
	})

	// Start writer goroutine
	go writePump(client)

	// Keep websocket alive
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			utils.ErrorLog.Println("❌ WS Error:", err)
			break
		}
	}
}

func writePump(client *service.WsClient) {

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(msg); err != nil {
				utils.ErrorLog.Println("❌ WS Error:", err)
				return
			}

		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			)

			if err := client.Conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				return
			}
		}
	}
}

// ============================================================
// DEVICE CONTROL
// ============================================================

func SendDeviceControl(deviceID string, device string, state int) {

	payload := fiber.Map{
		"device": device,
		"state":  state,
	}

	mqtt.SendCommand(deviceID, payload)
}

// ============================================================
// FAN
// ============================================================

func FanOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "fan", 1)

	return c.JSON(fiber.Map{
		"message": "fan on",
	})
}

func FanOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "fan", 0)

	return c.JSON(fiber.Map{
		"message": "fan off",
	})
}

// ============================================================
// MOTOR
// ============================================================

func MotorOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "motor", 1)

	return c.JSON(fiber.Map{
		"message": "motor on",
	})
}

func MotorOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "motor", 0)

	return c.JSON(fiber.Map{
		"message": "motor off",
	})
}

// ============================================================
// HEATER
// ============================================================

func HeaterOn(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "heater", 1)

	return c.JSON(fiber.Map{
		"message": "heater on",
	})
}

func HeaterOff(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	SendDeviceControl(deviceID, "heater", 0)

	return c.JSON(fiber.Map{
		"message": "heater off",
	})
}

// ============================================================
// MODE CONTROL
// ============================================================

func SetAutoMode(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	payload := fiber.Map{
		"method": "setMode",
		"params": true,
	}

	mqtt.SendCommand(deviceID, payload)

	stagepayload := fiber.Map{
		"method": "setStage",
		"params": 0,
	}

	mqtt.SendCommand(deviceID, stagepayload)

	return c.JSON(fiber.Map{
		"message": "AUTO mode enabled",
	})
}

func SetManualMode(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	payload := fiber.Map{
		"method": "setMode",
		"params": false,
	}

	mqtt.SendCommand(deviceID, payload)

	stagepayload := fiber.Map{
		"method": "setStage",
		"params": 0,
	}

	mqtt.SendCommand(deviceID, stagepayload)

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

func SetTempThreshold(c *fiber.Ctx) error {

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

	mqtt.SendCommand(deviceID, payload)

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

func SetHumThreshold(c *fiber.Ctx) error {

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

	mqtt.SendCommand(deviceID, payload)

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

func SetFanCycle(c *fiber.Ctx) error {

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

	mqtt.SendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "fan cycle updated",
	})
}

// Update stage settings

func SetStageSettings(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	var body []model.StageSettings

	if err := c.BodyParser(&body); err != nil {

		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	payload := fiber.Map{
		"method": "setStageSettings",
		"params": body,
	}

	mqtt.SendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "stage settings updated",
	})
}

// Set Stage

type SetStageRequest struct {
	Stage int `json:"stage"`
}

func SetStage(c *fiber.Ctx) error {

	deviceID := c.Params("id")

	var body SetStageRequest

	if err := c.BodyParser(&body); err != nil {

		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if body.Stage == 0 {
		payload := fiber.Map{
			"method": "setMode",
			"params": true,
		}

		mqtt.SendCommand(deviceID, payload)

		return c.JSON(fiber.Map{
			"message": "AUTO mode enabled",
		})
	}

	payload := fiber.Map{
		"method": "setStage",
		"params": body.Stage,
	}

	mqtt.SendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "stage updated",
	})
}

func GetAllTelemetry(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	limit, limiterr := strconv.Atoi(c.Query("limit"))
	if limiterr != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: limiterr.Error(),
			ErrorTime:    time.Now(),
		})
	}

	pagenumber, pagenumbererr := strconv.Atoi(c.Query("page"))
	if pagenumbererr != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: pagenumbererr.Error(),
			ErrorTime:    time.Now(),
		})
	}

	search := c.Query("search")

	offset := (pagenumber - 1) * limit

	//fetch data from DB
	result, count, err := service.GetAllTelemetry(ctx, int64(limit), int64(offset), search)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          &fiber.Map{"telemetry": result, "total_count": count},
	})
}
