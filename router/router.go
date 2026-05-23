package router

import (
	"sericulture/controller"
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
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	app.Use(logger.New())
	app.Use(recover.New())

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
	app.Post("/device/:id/temp-threshold", controller.SetTempThreshold)
	app.Post("/device/:id/hum-threshold", controller.SetHumThreshold)
	app.Post("/device/:id/fan-cycle", controller.SetFanCycle)
	app.Post("/device/:id/stage-settings", controller.SetStageSettings)
	app.Post("/device/:id/set-stage", controller.SetStage)

	app.Post("/user", controller.CreateUser)
	app.Put("/user/:id", controller.UpdateUser)
	app.Delete("/user/:id", controller.DeleteUser)
	app.Get("/users", controller.GetAllUsers)
	app.Get("/user/:id", controller.GetUserById)

	app.Post("/login", controller.Login)

	return app
}
