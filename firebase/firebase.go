package firebase

import (
	"context"
	"log"

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
		log.Fatal(err)
	}

	FirebaseClient = app

	log.Println("Firebase initialized successfully")
}
