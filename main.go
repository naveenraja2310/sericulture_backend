package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sericulture/database"
	"sericulture/router"
	"sericulture/settings"
	"syscall"

	"sericulture/mqtt"
)

func main() {
	config, err := settings.InitConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if dberr := database.InitDB(config); dberr != nil {
		log.Fatalf("Failed to initialize database: %v", dberr)
	}

	mqtt.ConnectMQTT(config.MQTTURL)

	router := router.GetRouter()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Listen(fmt.Sprintf(":%s", config.AppPort)); err != nil {
			log.Printf("Failed to start server: %v", err)
		}
	}()

	<-quit

	log.Println("Shutting down server...")
}
