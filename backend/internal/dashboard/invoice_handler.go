package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// InvoiceData holds one invoice's information plus a list of associated payments.
type InvoiceData struct {
	InvoiceID      string        `json:"invoiceID"` // The internal QBO invoice ID
	DocNumber      string        `json:"docNumber"` // The invoice # shown to the parent
	BillEmail      string        `json:"billEmail"`
	CreatedTime    string        `json:"createdTime"`
	LastUpdated    string        `json:"lastUpdated"`
	Balance        interface{}   `json:"balance,omitempty"`        // Could be float or missing
	HoursPurchased interface{}   `json:"hoursPurchased,omitempty"` // Could be float or missing
	IsVoided       bool          `json:"isVoided"`
	Payments       []PaymentData `json:"payments"`
}

// PaymentData holds the fields for a single payment document.
type PaymentData struct {
	PaymentID          string      `json:"paymentID"`
	CreatedAt          string      `json:"created_at"`
	PaymentOnInvoice   string      `json:"payment_on_invoice"`   // Now always a string
	TotalPaymentAmount interface{} `json:"total_payment_amount"` // Could be float or missing
	PaymentMethod      string      `json:"paymentMethod"`
}

// ParentInvoicesResponse is the overall JSON returned to the client.
type ParentInvoicesResponse struct {
	ParentID string        `json:"parentID"`
	Invoices []InvoiceData `json:"invoices"`
}

// GetParentInvoicesHandler fetches all invoice docs (and payments sub-docs) for a given parent.
func (a *App) GetParentInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Identify parent from JWT
	parentID, _ := a.getParentCredentials(r)
	if parentID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	// 2. Get parent's doc in 'parents' collection
	ctx := context.Background()
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(parentID)
	parentSnap, err := parentDocRef.Get(ctx)
	if err != nil || !parentSnap.Exists() {
		log.Printf("Error or missing parent doc for user %s: %v", parentID, err)
		http.Error(w, "Parent document not found", http.StatusNotFound)
		return
	}
	parentData := parentSnap.Data()

	// Extract qboCustomerId from parent's business sub-doc
	businessData, ok := parentData["business"].(map[string]interface{})
	if !ok {
		http.Error(w, "No business sub-doc found for parent", http.StatusBadRequest)
		return
	}
	qboCustomerID, ok := businessData["qboCustomerId"].(string)
	if !ok || qboCustomerID == "" {
		http.Error(w, "Invalid or missing qboCustomerId in parent's business sub-doc", http.StatusBadRequest)
		return
	}

	// 3. Access intuit/{qboCustomerId} doc; get 'invoices' subcollection
	intuitDocRef := a.FirestoreClient.Collection("intuit").Doc(qboCustomerID)
	invoicesRef := intuitDocRef.Collection("invoices")

	// Fetch all invoice docs
	invoiceDocs, err := invoicesRef.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching invoices subcollection for parent %s: %v", parentID, err)
		http.Error(w, "Failed to fetch invoices", http.StatusInternalServerError)
		return
	}

	var allInvoices []InvoiceData

	// 4. Iterate over invoice docs
	for _, invDoc := range invoiceDocs {
		docData := invDoc.Data()

		// invoiceID and docNumber as strings
		invoiceID, _ := docData["invoiceID"].(string)
		docNumber, _ := docData["docNumber"].(string)
		billEmail, _ := docData["billEmail"].(string)

		// Attempt time.Time or string for createdTime
		var createdTimeStr string
		if t, ok := docData["createdTime"].(time.Time); ok {
			createdTimeStr = t.Format(time.RFC3339)
		} else if st, ok := docData["createdTime"].(string); ok {
			createdTimeStr = st
		} else {
			createdTimeStr = ""
		}

		// Attempt time.Time or string for lastUpdated
		var lastUpdatedStr string
		if t, ok := docData["lastUpdated"].(time.Time); ok {
			lastUpdatedStr = t.Format(time.RFC3339)
		} else if st, ok := docData["lastUpdated"].(string); ok {
			lastUpdatedStr = st
		} else {
			lastUpdatedStr = ""
		}

		// We can store balance/hours as interface{} to preserve float or "missing"
		balance, hasBalance := docData["balance"]
		hoursPurchased, hasHours := docData["hoursPurchased"]

		// If both are missing => invoice is voided
		isVoided := (!hasBalance && !hasHours)

		invoice := InvoiceData{
			InvoiceID:      invoiceID,
			DocNumber:      docNumber,
			BillEmail:      billEmail,
			CreatedTime:    createdTimeStr,
			LastUpdated:    lastUpdatedStr,
			Balance:        balance,
			HoursPurchased: hoursPurchased,
			IsVoided:       isVoided,
			Payments:       []PaymentData{},
		}

		// 5. For each invoice doc, fetch 'payments' subcollection
		paymentsRef := invDoc.Ref.Collection("payments")
		payDocs, err := paymentsRef.Documents(ctx).GetAll()
		if err != nil {
			log.Printf("Error fetching payments subcollection for invoice %s: %v", invoiceID, err)
			// We continue, but payments slice remains empty
		} else {
			for _, payDoc := range payDocs {
				payData := payDoc.Data()

				paymentID, _ := payData["paymentID"].(string)
				paymentMethod, _ := payData["paymentMethod"].(string)

				// Parse created_at as time or string
				var createdAtStr string
				if pt, ok := payData["created_at"].(time.Time); ok {
					createdAtStr = pt.Format(time.RFC3339)
				} else if st, ok := payData["created_at"].(string); ok {
					createdAtStr = st
				} else {
					createdAtStr = ""
				}

				// PaymentOnInvoice can be a float, int, or string => always convert to string for JSON
				paymentOnInvoiceStr := fmt.Sprintf("%v", payData["payment_on_invoice"])

				// totalPaymentAmount can be any type
				totalPaymentAmount, _ := payData["total_payment_amount"]

				payment := PaymentData{
					PaymentID:          paymentID,
					CreatedAt:          createdAtStr,
					PaymentOnInvoice:   paymentOnInvoiceStr,
					TotalPaymentAmount: totalPaymentAmount,
					PaymentMethod:      paymentMethod,
				}
				invoice.Payments = append(invoice.Payments, payment)
			}
		}

		allInvoices = append(allInvoices, invoice)
	}

	// 6. Build final response
	resp := ParentInvoicesResponse{
		ParentID: parentID,
		Invoices: allInvoices,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding invoice response for parent %s: %v", parentID, err)
		http.Error(w, "Failed to encode invoice response", http.StatusInternalServerError)
	}
}
