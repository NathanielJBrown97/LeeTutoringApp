package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type FamilyBillingData struct {
	TotalHours   float64 `json:"total_hours"`
	TotalBalance float64 `json:"total_balance"`
}

func (a *App) TotalHoursAndBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := a.getParentCredentials(r)
	if userID == "" {
		http.Error(w, "The parent's User ID was not found...", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(userID)
	parentDocSnap, err := parentDocRef.Get(ctx)
	if err != nil {
		log.Printf("Error fetchign parent doc for userID %s: %v", userID, err)
		http.Error(w, "Failed to fetch parent data", http.StatusInternalServerError)
		return
	}
	if !parentDocSnap.Exists() {
		log.Printf("Parent document does not exist for userID %s", userID)
		http.Error(w, "Parent data not found", http.StatusNotFound)
		return
	}

	parentBusinessData := parentDocSnap.Data()
	businessField, ok := parentBusinessData["business"].(map[string]interface{})
	if !ok {
		log.Printf("no business sub doc found for user id %s", userID)
		http.Error(w, "business data not foudn for parent", http.StatusNotFound)
		return
	}
	qboCustomerId, ok := businessField["qboCustomerId"].(string)
	if !ok || qboCustomerId == "" {
		log.Printf("No valid qbocustomer id in business sub doc for userID %s", userID)
		http.Error(w, "qboCustomerId not found in businessfield", http.StatusNotFound)
		return
	}
	intuitDocRef := a.FirestoreClient.Collection("intuit").Doc(qboCustomerId)
	intuitDocSnap, err := intuitDocRef.Get(ctx)
	if err != nil {
		log.Printf("error fetching intuit doc for qbocustomerid %s: %v", qboCustomerId, err)
		http.Error(w, "failed to fetch intuit data", http.StatusInternalServerError)
		return
	}
	if !intuitDocSnap.Exists() {
		log.Printf("Intuit document does not exist for qboCustomer ID %s", qboCustomerId)
		http.Error(w, "intuit data not found", http.StatusNotFound)
		return
	}
	intuitData := intuitDocSnap.Data()
	totalBalance, _ := intuitData["total_balance"].(float64)
	totalHours, _ := intuitData["total_hours"].(float64)

	response := FamilyBillingData{
		TotalBalance: totalBalance,
		TotalHours:   totalHours,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
