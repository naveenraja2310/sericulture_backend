package firebase

import (
	"context"
	"sericulture/utils"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
)

var (
	FirebaseClient *firebase.App
)

func InitFirebase() {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "./firebase-service-account.json")

	config := &firebase.Config{
		ProjectID: "yadhronics",
	}

	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		utils.Error(nil, "Failed to initialize Firebase", "error", err.Error())
		return
	}

	FirebaseClient = app
	utils.Info(nil, "Firebase initialized successfully")
}
