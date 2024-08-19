package config

import (
	"context"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitializeFirebase(ctx context.Context) (*firebase.App, error) {
	// Path to the service account key file
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}
	return app, nil
}
