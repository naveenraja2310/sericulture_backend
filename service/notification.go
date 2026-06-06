package service

import (
	"context"
	"sericulture/database"
	"sericulture/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllNotification(ctx context.Context, limit, offset int64, deviceID string) ([]model.Notification, int64, error) {
	var notification []model.Notification

	// Define find options for pagination and sorting
	filter := bson.M{"deviceId": deviceID} // Filter by device ID

	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by created_at field, descending

	// Apply limit if it's greater than 0
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	// Apply offset for pagination
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := database.Notification.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode the notification from the cursor
	if err = cursor.All(ctx, &notification); err != nil {
		return nil, 0, err
	}

	count, err := database.Notification.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return notification, count, nil
}
