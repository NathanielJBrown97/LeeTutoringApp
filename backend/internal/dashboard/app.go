// backend/internal/dashboard/app.go

package dashboard

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
)

// App holds the dependencies for the dashboard package
type App struct {
	Config          *config.Config
	FirestoreClient *firestore.Client
}
