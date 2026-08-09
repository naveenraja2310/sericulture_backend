package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sericulture/database"
	"sericulture/firebase"
	"sericulture/router"
	"sericulture/settings"
	"syscall"
	"time"

	"sericulture/mqtt"
	"sericulture/utils"
)

func main() {
	log.SetOutput(os.Stdout)

	config, err := settings.InitConfig()
	if err != nil {
		utils.Error(nil, "Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	if dberr := database.InitDB(config); dberr != nil {
		utils.Error(nil, "Failed to initialize database", "error", dberr.Error())
		os.Exit(1)
	}

	if config.GrafanaInstanceID != "" && config.GrafanaAPIKey != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		provider, grafanaErr := utils.InitGrafanaOTLPLogger(ctx, config.GrafanaInstanceID, config.GrafanaAPIKey)
		cancel()
		if grafanaErr != nil {
			utils.Error(nil, "Failed to initialize Grafana OTLP logger", "error", grafanaErr.Error())
		} else if provider != nil {
			utils.Info(nil, "Grafana OTLP logger initialized successfully")
		}
	}

	firebase.InitFirebase()

	mqtt.ConnectMQTT(config.MQTTURL)

	router := router.GetRouter()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Listen(fmt.Sprintf(":%s", config.AppPort)); err != nil {
			utils.Error(nil, "Failed to start server", "error", err.Error())
		}
	}()

	<-quit

	utils.Info(nil, "Shutting down server...")
}
