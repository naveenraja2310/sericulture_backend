package utils

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func StringToObjectID(hex string) primitive.ObjectID {
	objID, _ := primitive.ObjectIDFromHex(hex)
	return objID
}

/*
IsDuplicate checks if a document matching the given query already exists in the specified collection.
It returns true if a matching document is found, otherwise false.
*/
func IsDuplicate(ctx context.Context, collection *mongo.Collection, query interface{}) bool {
	// Perform a query to find a matching document in the collection.
	result := map[string]interface{}{}
	err := collection.FindOne(ctx, query).Decode(result)
	if err != nil {
		// Log the absence of a duplicate or any errors encountered during the search.
		if err == mongo.ErrNoDocuments {
			return false
		}
		return false
	}
	return true
}

/*
CheckIfExistsByID checks if a document with the given ID exists in the specified collection.
It returns an error if the document is not found or if there's an issue during the search.
*/
func CheckIfExistsByID(ctx context.Context, collection *mongo.Collection, id primitive.ObjectID) bool {
	// Attempt to find the document in the specified collection.
	result := map[string]interface{}{}
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
	}

	return true
}
