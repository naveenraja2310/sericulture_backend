package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sericulture/database"
	"sericulture/firebase"
	"sericulture/router"
	"sericulture/settings"
	"syscall"

	"sericulture/mqtt"
	"sericulture/utils"
)

func main() {
	log.SetOutput(os.Stdout)

	config, err := settings.InitConfig()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to load configuration: %v", err)
	}

	if dberr := database.InitDB(config); dberr != nil {
		utils.ErrorLog.Fatalf("Failed to initialize database: %v", dberr)
	}

	firebase.InitFirebase()

	mqtt.ConnectMQTT(config.MQTTURL)

	router := router.GetRouter()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Listen(fmt.Sprintf(":%s", config.AppPort)); err != nil {
			utils.ErrorLog.Printf("Failed to start server: %v", err)
		}
	}()

	<-quit

	log.Println("Shutting down server...")
}
