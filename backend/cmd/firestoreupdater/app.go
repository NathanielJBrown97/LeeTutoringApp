package firestoreupdater

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
)

type App struct {
	Config          *config.Config
	FirestoreClient *firestore.Client
}
