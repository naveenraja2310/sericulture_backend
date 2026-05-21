package service

import (
	"context"
	"errors"
	"sericulture/database"
	"sericulture/model"
	"sericulture/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Login(ctx context.Context, username string, password string) (model.User, error) {
	var user model.User
	err := database.Users.FindOne(ctx, bson.M{"username": username, "password": password}).Decode(&user)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	user.CreatedAt = time.Now()

	_, err := database.Users.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(ctx context.Context, user model.User, id primitive.ObjectID) (*model.User, error) {
	if !utils.CheckIfExistsByID(ctx, database.Users, id) {
		return nil, errors.New("the given id is invalid")
	}

	updateFields := bson.M{"updatedAt": time.Now()}

	updateFields["username"] = user.Username
	updateFields["password"] = user.Password
	updateFields["deviceId"] = user.DeviceID

	update := bson.M{"$set": updateFields}
	result, err := database.Users.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("no document found with the given id")
	}

	return &user, nil
}

func UpdateUserAndDeviceId(ctx context.Context, user model.User, id primitive.ObjectID) (*model.User, error) {
	if !utils.CheckIfExistsByID(ctx, database.Users, id) {
		return nil, errors.New("the given id is invalid")
	}

	updateFields := bson.M{"updatedAt": time.Now()}

	if user.Username != "" {
		updateFields["username"] = user.Username
	}
	if user.Password != "" {
		updateFields["password"] = user.Password
	}
	if user.DeviceID != "" {
		updateFields["deviceId"] = user.DeviceID
	}

	update := bson.M{"$set": updateFields}
	result, err := database.Users.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("no document found with the given id")
	}

	return &user, nil
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User

	if !utils.CheckIfExistsByID(ctx, database.Users, id) {
		return nil, errors.New("the given id is invalid")
	}

	err := database.Users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	if !utils.CheckIfExistsByID(ctx, database.Users, id) {
		return errors.New("the given id is invalid")
	}

	result, err := database.Users.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no document found with the given id")
	}

	return nil
}

func GetAllUsers(ctx context.Context, limit, offset int64, search string) ([]model.User, int64, error) {
	var users []model.User

	// Define find options for pagination and sorting
	filter := bson.M{}

	if search != "" {
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by created_at field, descending

	// Apply limit if it's greater than 0
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	// Apply offset for pagination
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := database.Users.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode the users from the cursor
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	count, err := database.Users.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}
