// backend/internal/config/config.go

package config

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type Config struct {
	FIREBASE_PROJECT_ID            string
	GOOGLE_CLIENT_ID               string
	GOOGLE_CLIENT_SECRET           string
	GOOGLE_REDIRECT_URL            string
	MICROSOFT_CLIENT_ID            string
	MICROSOFT_CLIENT_SECRET        string
	MICROSOFT_REDIRECT_URL         string
	MICROSOFT_TENANT_ID            string
	YAHOO_CLIENT_ID                string
	YAHOO_CLIENT_SECRET            string
	YAHOO_REDIRECT_URL             string
	SESSION_SECRET                 string
	SECRET_KEY                     string
	GOOGLE_CLOUD_PROJECT           string
	GOOGLE_APPLICATION_CREDENTIALS string
}

func LoadConfig() (*Config, error) {
	return &Config{
		FIREBASE_PROJECT_ID:            os.Getenv("FIREBASE_PROJECT_ID"),
		GOOGLE_CLIENT_ID:               os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET:           os.Getenv("GOOGLE_CLIENT_SECRET"),
		GOOGLE_REDIRECT_URL:            os.Getenv("GOOGLE_REDIRECT_URL"),
		MICROSOFT_CLIENT_ID:            os.Getenv("MICROSOFT_CLIENT_ID"),
		MICROSOFT_CLIENT_SECRET:        os.Getenv("MICROSOFT_CLIENT_SECRET"),
		MICROSOFT_REDIRECT_URL:         os.Getenv("MICROSOFT_REDIRECT_URL"),
		MICROSOFT_TENANT_ID:            os.Getenv("MICROSOFT_TENANT_ID"),
		YAHOO_CLIENT_ID:                os.Getenv("YAHOO_CLIENT_ID"),
		YAHOO_CLIENT_SECRET:            os.Getenv("YAHOO_CLIENT_SECRET"),
		YAHOO_REDIRECT_URL:             os.Getenv("YAHOO_REDIRECT_URL"),
		SESSION_SECRET:                 os.Getenv("SESSION_SECRET"),
		SECRET_KEY:                     os.Getenv("SECRET_KEY"),
		GOOGLE_CLOUD_PROJECT:           os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GOOGLE_APPLICATION_CREDENTIALS: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}, nil
}

func InitializeFirestore(cfg *Config) (*firestore.Client, error) {
	ctx := context.Background()

	// Use the credentials file specified by GOOGLE_APPLICATION_CREDENTIALS
	clientOptions := option.WithCredentialsFile(cfg.GOOGLE_APPLICATION_CREDENTIALS)

	// Initialize the Firestore client with the project ID and credentials
	client, err := firestore.NewClient(ctx, cfg.GOOGLE_CLOUD_PROJECT, clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}
