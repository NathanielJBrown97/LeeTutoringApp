// backend/internal/config/config.go

package config

import (
	"fmt"
	"log"

	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Config holds all the configuration values
type Config struct {
	// Firebase configurations
	FIREBASE_PROJECT_ID       string `env:"FIREBASE_PROJECT_ID,required"`
	FIREBASE_CREDENTIALS_FILE string `env:"FIREBASE_CREDENTIALS_FILE,required"`

	// Session configurations
	SESSION_SECRET string `env:"SESSION_SECRET,required"`

	// Server configurations
	PORT string `env:"PORT" envDefault:"8080"`

	// Google OAuth2 configurations
	GOOGLE_CLIENT_ID     string `env:"GOOGLE_CLIENT_ID,required"`
	GOOGLE_CLIENT_SECRET string `env:"GOOGLE_CLIENT_SECRET,required"`
	GOOGLE_REDIRECT_URL  string `env:"GOOGLE_REDIRECT_URL,required"`

	// Microsoft OAuth2 configurations
	MICROSOFT_CLIENT_ID     string `env:"MICROSOFT_CLIENT_ID,required"`
	MICROSOFT_CLIENT_SECRET string `env:"MICROSOFT_CLIENT_SECRET,required"`
	MICROSOFT_REDIRECT_URL  string `env:"MICROSOFT_REDIRECT_URL,required"`
}

// LoadConfig loads environment variables and returns a Config struct
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error reading .env file, proceeding with environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Custom validation if needed
	if cfg.FIREBASE_PROJECT_ID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID is required but not set")
	}
	if cfg.FIREBASE_CREDENTIALS_FILE == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_FILE is required but not set")
	}
	if cfg.SESSION_SECRET == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required but not set")
	}
	if cfg.GOOGLE_CLIENT_ID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID is required but not set")
	}
	if cfg.GOOGLE_CLIENT_SECRET == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET is required but not set")
	}
	if cfg.GOOGLE_REDIRECT_URL == "" {
		return nil, fmt.Errorf("GOOGLE_REDIRECT_URL is required but not set")
	}
	if cfg.MICROSOFT_CLIENT_ID == "" {
		return nil, fmt.Errorf("MICROSOFT_CLIENT_ID is required but not set")
	}
	if cfg.MICROSOFT_CLIENT_SECRET == "" {
		return nil, fmt.Errorf("MICROSOFT_CLIENT_SECRET is required but not set")
	}
	if cfg.MICROSOFT_REDIRECT_URL == "" {
		return nil, fmt.Errorf("MICROSOFT_REDIRECT_URL is required but not set")
	}

	return cfg, nil
}

// InitializeFirestore initializes and returns a Firestore client using the configuration
func InitializeFirestore(cfg *Config) (*firestore.Client, error) {
	ctx := context.Background()

	// Load the service account key file from the path specified in the config
	serviceAccount := option.WithCredentialsFile(cfg.FIREBASE_CREDENTIALS_FILE)

	// Initialize Firebase app with project ID
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: cfg.FIREBASE_PROJECT_ID,
	}, serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Create a Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return client, nil
}
