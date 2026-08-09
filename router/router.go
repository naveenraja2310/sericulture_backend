package router

import (
	"sericulture/controller"
	"sericulture/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

func GetRouter() *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(utils.JWTMiddleware)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/device/:id/status", controller.GetDeviceStatus)

	app.Get("/device/:id/ws",
		func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("deviceID", c.Params("id"))
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		},
		websocket.New(controller.DeviceStatusWS, websocket.Config{
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: true,
		}),
	)

	app.Post("/device/:id/fan/on", controller.FanOn)
	app.Post("/device/:id/fan/off", controller.FanOff)
	app.Post("/device/:id/motor/on", controller.MotorOn)
	app.Post("/device/:id/motor/off", controller.MotorOff)
	app.Post("/device/:id/heater/on", controller.HeaterOn)
	app.Post("/device/:id/heater/off", controller.HeaterOff)
	app.Post("/device/:id/mode/auto", controller.SetAutoMode)
	app.Post("/device/:id/mode/manual", controller.SetManualMode)
	app.Post("/device/:id/threshold", controller.SetThreshold)
	app.Post("/device/:id/stage-settings", controller.SetStageSettings)
	app.Post("/device/:id/set-stage", controller.SetStage)
	app.Post("/device/:id/system-enabled", controller.SetSystemEnabled)
	app.Get("/get-all-telemetry", controller.GetAllTelemetry)

	app.Post("/user", controller.CreateUser)
	app.Put("/user/:id", controller.UpdateUser)
	app.Delete("/user/:id", controller.DeleteUser)
	app.Get("/users", controller.GetAllUsers)
	app.Get("/user/:id", controller.GetUserById)

	app.Get("/notification", controller.GetAllNotification)

	app.Post("/upload", controller.UploadFile)
	app.Get("/file/:filename", controller.GetFile)
	app.Post("/update-firmware/:id", controller.UpdateFirmware)

	app.Post("/login", controller.Login)
	app.Put("/logout/:id", controller.LogOut)

	return app
}
