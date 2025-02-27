package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

// UpdateInvoiceEmailRequest represents the JSON payload for updating the invoice email.
type UpdateInvoiceEmailRequest struct {
	ParentID     string `json:"parent_id"`     // The parent's document ID
	InvoiceEmail string `json:"invoice_email"` // The new invoice email address
}

// UpdateInvoiceEmailHandler handles POST requests to update the parent's invoice email.
// It updates (or creates) the field "business.invoice_email" in the parent's document.
func (app *App) UpdateInvoiceEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed; use POST", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateInvoiceEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Try to update the field in Firestore.
	_, err := app.FirestoreClient.Collection("parents").Doc(req.ParentID).Update(ctx, []firestore.Update{
		{Path: "business.invoice_email", Value: req.InvoiceEmail},
	})
	if err != nil {
		// Log the detailed error so you can see what Firestore is complaining about.
		log.Printf("Error updating invoice_email for parent %s: %v", req.ParentID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error while updating",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// GetInvoiceEmailHandler handles GET requests to fetch the parent's invoice email.
// It expects the parent's ID as a query parameter (e.g., ?parent_id=abc123).
func (app *App) GetInvoiceEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed; use GET", http.StatusMethodNotAllowed)
		return
	}

	parentID := r.URL.Query().Get("parent_id")
	if parentID == "" {
		http.Error(w, "Missing parent_id parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	doc, err := app.FirestoreClient.Collection("parents").Doc(parentID).Get(ctx)
	if err != nil {
		http.Error(w, "Error retrieving parent document", http.StatusInternalServerError)
		return
	}

	data := doc.Data()
	invoiceEmail := ""
	if business, ok := data["business"].(map[string]interface{}); ok {
		if email, exists := business["invoice_email"].(string); exists {
			invoiceEmail = email
		}
	}

	resp := map[string]string{
		"invoice_email": invoiceEmail,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
