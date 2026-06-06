package database

import (
	"context"
	"log"
	"sericulture/settings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient                    *mongo.Client
	Users, Telemetry, Notification *mongo.Collection
	ContextTime                    int = 10
)

func InitDB(config settings.Configuration) error {
	clientOptions := options.Client().ApplyURI(config.DBURI)

	// Connect to MongoDB using the client options
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	Users = client.Database(config.DB_NAME).Collection("users")
	Telemetry = client.Database(config.DB_NAME).Collection("telemetry")
	Notification = client.Database(config.DB_NAME).Collection("notification")

	log.Println("Database loaded successfully")
	return nil
}
