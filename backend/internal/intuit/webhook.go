package intuit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2"
)

// A sample JSON structure for the QuickBooks Webhook payload
// Reference: https://developer.intuit.com/app/developer/qbo/docs/develop/webhooks
type WebhookEvent struct {
	EventNotifications []struct {
		RealmID         string `json:"realmId"`
		DataChangeEvent struct {
			Entities []struct {
				Name            string `json:"name"`
				ID              string `json:"id"`
				Operation       string `json:"operation"`
				LastUpdatedTime string `json:"lastUpdatedTime"`
			} `json:"entities"`
		} `json:"dataChangeEvent"`
	} `json:"eventNotifications"`
}

// Minimal struct for storing invoice data in Firestore
type InvoiceRecord struct {
	InvoiceID      string    `firestore:"invoiceID,omitempty"` // the QBO internal ID
	DocNumber      string    `firestore:"docNumber,omitempty"` // the user-facing "Invoice #"
	CustomerRef    string    `firestore:"customerRef,omitempty"`
	BillEmail      string    `firestore:"billEmail,omitempty"`
	CreatedTime    time.Time `firestore:"createdTime,omitempty"`
	LastUpdated    time.Time `firestore:"lastUpdated,omitempty"`
	Balance        float64   `firestore:"balance,omitempty"`
	HoursPurchased float64   `firestore:"hoursPurchased,omitempty"`
}

// HandleWebhook is the HTTP endpoint your QuickBooks webhook calls when an
// invoice is created/updated in the QBO UI. You must configure this endpoint
// in the Intuit Developer Portal under your app's "Webhooks" settings.
func (s *OAuthService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Optional: Validate the webhook signature from Intuit ("intuit-signature" header).
	// See QBO docs for proper signature verification. We'll skip it for brevity.

	// 2. Parse the JSON payload
	var webhook WebhookEvent
	if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// 3. For each event notification, check the realmId and the entities
	for _, note := range webhook.EventNotifications {
		realmID := note.RealmID

		// The dataChangeEvent -> entities. They might be multiple.
		for _, entity := range note.DataChangeEvent.Entities {

			log.Printf("[Webhook Debug] entity.Name=%s, entity.ID=%s, entity.Operation=%s",
				entity.Name, entity.ID, entity.Operation)
			// Check if it's an Invoice event
			if strings.EqualFold(entity.Name, "Invoice") {
				invoiceID := entity.ID
				operation := entity.Operation

				if strings.EqualFold(operation, "Delete") ||
					strings.EqualFold(operation, "Remove") ||
					strings.EqualFold(operation, "Deleted") ||
					strings.EqualFold(operation, "Removed") {
					// We skip fetching from QBO because the invoice is gone
					err := s.deleteInvoiceDoc(r.Context(), invoiceID)
					if err != nil {
						log.Printf("Failed to delete invoice doc %s: %v\n", invoiceID, err)
					} else {
						log.Printf("Webhook: successfully handled delete event for invoice %s\n", invoiceID)
					}
					continue
				}

				// otherwise handle create/update/void
				invData, err := s.fetchInvoiceFromQBO(r.Context(), realmID, invoiceID)
				if err != nil {
					log.Printf("Failed to fetch invoice %s from realm %s: %v\n", invoiceID, realmID, err)
					continue
				}

				// Store it in Firestore: intuit/<customerRef>/invoices/<invoiceID>
				err = s.storeInvoice(r.Context(), invData)
				if err != nil {
					log.Printf("Failed to store invoice in Firestore: %v\n", err)
				} else {
					log.Printf("Webhook: successfully handled %s event for invoice %s\n", operation, invoiceID)
				}
			}
		}
	}

	// Respond 200 to acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

func (s *OAuthService) deleteInvoiceDoc(ctx context.Context, invoiceID string) error {
	//   Look for doc where invoiceId == invoiceID
	log.Printf("deleteInvoiceDoc called for invoiceID=%s", invoiceID)
	colRef := s.firestore.CollectionGroup("invoices").
		Where("invoiceID", "==", invoiceID).
		Limit(1)

	snaps, err := colRef.Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(snaps) == 0 {
		// doc not found, maybe it's already deleted
		return nil
	}

	docRef := snaps[0].Ref
	if _, err := docRef.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete invoice doc %s: %w", invoiceID, err)
	}
	return nil
}

func (s *OAuthService) fetchCustomerFromQBO(ctx context.Context, realmID, customerID string) (string, error) {
	// Reuse your getGlobalTokens -> refreshAccessTokenIfExpired logic
	token, err := s.getGlobalTokens(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tokens: %w", err)
	}
	validToken, err := s.refreshAccessTokenIfExpired(ctx, token, realmID)
	if err != nil {
		return "", fmt.Errorf("refresh token error: %w", err)
	}

	client := s.config.Client(ctx, validToken)
	url := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/customer/%s?minorversion=65", realmID, customerID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create customer request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("customer fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("customer fetch returned status %d", resp.StatusCode)
	}

	// Minimal structure for a QBO customer
	var custResp struct {
		Customer struct {
			Id               string `json:"Id"`
			PrimaryEmailAddr string `json:"PrimaryEmailAddr"`
			// or "PrimaryEmailAddr": { "Address": "..."} if QBO returns it as an object
		} `json:"Customer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&custResp); err != nil {
		return "", fmt.Errorf("failed to decode customer JSON: %w", err)
	}

	return custResp.Customer.PrimaryEmailAddr, nil
}

// fetchInvoiceFromQBO retrieves the full invoice details via the QuickBooks API.
// We'll also refresh the token if needed.
func (s *OAuthService) fetchInvoiceFromQBO(ctx context.Context, realmID, invoiceID string) (*InvoiceRecord, error) {
	// 1. Retrieve your stored OAuth tokens from somewhere. If you do only
	//    one realm for your whole business, you might store them in some
	//    known doc, or you have them in the parent's doc. For now, let's
	//    assume you have them at s.getGlobalTokens(...) or similar.
	token, err := s.getGlobalTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global tokens: %w", err)
	}

	// 2. Refresh if expired
	validToken, err := s.refreshAccessTokenIfExpired(ctx, token, realmID)
	if err != nil {
		return nil, fmt.Errorf("refresh token error: %w", err)
	}

	// 3. Build the request to QBO API
	client := s.config.Client(ctx, validToken)
	url := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/invoice/%s?minorversion=65", realmID, invoiceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build invoice fetch request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	// 4. Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("invoice fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invoice fetch returned status %d", resp.StatusCode)
	}

	// 5. Parse the invoice JSON
	var invoiceResp struct {
		Invoice struct {
			ID          string `json:"Id"`
			DocNumber   string `json:"DocNumber"`
			CustomerRef struct {
				Value string `json:"value"`
			} `json:"CustomerRef"`
			// BillEmail to collect email address
			BillEmail struct {
				Address string `json:"Address"`
			} `json:"BillEmail"`
			// "TxnDate", "DueDate", etc., if you want them
			Balance  float64 `json:"Balance"`
			MetaData struct {
				CreateTime      string `json:"CreateTime"`
				LastUpdatedTime string `json:"LastUpdatedTime"`
			} `json:"MetaData"`
			Line []struct {
				DetailType          string  `json:"DetailType"`
				Amount              float64 `json:"Amount"`
				SalesItemLineDetail struct {
					Qty float64 `json:"Qty"`
					// Possibly UnitPrice, etc.
				} `json:"SalesItemLineDetail,omitempty"`
			} `json:"Line"`
		} `json:"Invoice"`
		Time string `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to decode invoice JSON: %w", err)
	}

	// 6. Extract relevant fields for Firestore
	invoice := invoiceResp.Invoice
	docNumber := invoice.DocNumber
	createTime, _ := time.Parse(time.RFC3339, invoice.MetaData.CreateTime)
	updateTime, _ := time.Parse(time.RFC3339, invoice.MetaData.LastUpdatedTime)

	// Attempt to parse BillEmail
	billEmail := invoice.BillEmail.Address

	// If it's empty, fetch from the Customer record
	if billEmail == "" {
		log.Printf("Invoice %s in realm %s has no BillEmail, fetching customer record...\n", invoiceID, realmID)
		custEmail, err2 := s.fetchCustomerFromQBO(ctx, realmID, invoice.CustomerRef.Value)
		if err2 != nil {
			log.Printf("Failed to fetch Customer for invoice %s: %v\n", invoiceID, err2)
		} else {
			billEmail = custEmail
			log.Printf("Got email '%s' from Customer record\n", billEmail)
		}
	}

	var totalHours float64
	for _, line := range invoice.Line {
		if line.DetailType == "SalesItemLineDetail" {
			totalHours += line.SalesItemLineDetail.Qty
		}
	}

	invRecord := &InvoiceRecord{
		InvoiceID:      invoice.ID,
		DocNumber:      docNumber,
		CustomerRef:    invoice.CustomerRef.Value,
		BillEmail:      billEmail,
		CreatedTime:    createTime,
		LastUpdated:    updateTime,
		Balance:        invoice.Balance,
		HoursPurchased: totalHours,
	}

	return invRecord, nil
}

// storeInvoice writes the invoice to Firestore under:
// `intuit/<customerRef>/invoices/<invoiceId>` doc.
func (s *OAuthService) storeInvoice(ctx context.Context, inv *InvoiceRecord) error {
	if inv == nil {
		return fmt.Errorf("invoice is nil")
	}
	if inv.CustomerRef == "" {
		return fmt.Errorf("missing customerRef in invoice data")
	}
	if inv.InvoiceID == "" {
		return fmt.Errorf("missing invoiceID in invoice data")
	}

	docRef := s.firestore.Collection("intuit").
		Doc(inv.CustomerRef).
		Collection("invoices").
		Doc(inv.InvoiceID)

	_, err := docRef.Set(ctx, inv)
	if err != nil {
		return fmt.Errorf("failed to store invoice record: %w", err)
	}

	//auto association
	if err := s.autoAssociateParent(ctx, inv); err != nil {
		log.Printf("AutoAssociateParent fucntion error: %v\n", err)
	}

	return nil
}

func (s *OAuthService) autoAssociateParent(ctx context.Context, inv *InvoiceRecord) error {
	if inv.BillEmail == "" {
		return nil
	}

	// 1. If there's already a doc with business.qboCustomerId == inv.CustomerRef, skip
	alreadyLinked, err := s.findParentWithQboId(ctx, inv.CustomerRef)
	if err != nil {
		return err
	}
	if alreadyLinked {
		log.Printf("Skipping re-assign: qboCustomerId=%s is already linked\n", inv.CustomerRef)
		return nil
	}

	// 2. Find parent doc by email
	parentsSnap, err := s.firestore.Collection("parents").
		Where("email", "==", inv.BillEmail).
		Limit(1).
		Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(parentsSnap) == 0 {
		return nil // no parent with that email
	}

	parentDoc := parentsSnap[0]

	// 3. Read parent's "business" sub-document (a map) if it exists
	// We'll do a normal .Data() read, then parse out "business" if it exists
	pData := parentDoc.Data()

	var businessData map[string]interface{}
	if rawBiz, ok := pData["business"]; ok {
		if casted, ok2 := rawBiz.(map[string]interface{}); ok2 {
			businessData = casted
		}
	} else {
		businessData = make(map[string]interface{}) // create a new one if not present
	}

	qboId, _ := businessData["qboCustomerId"].(string)

	// 4. If empty, set it
	if qboId == "" {
		businessData["qboCustomerId"] = inv.CustomerRef
		// update parent doc with new "business" map
		_, err := parentDoc.Ref.Set(ctx, map[string]interface{}{
			"business": businessData,
		}, firestore.MergeAll)
		if err != nil {
			return fmt.Errorf("failed to set qboCustomerId in parent doc sub-document: %w", err)
		}
		log.Printf("Auto-associated parent doc %s with qboCustomerId=%s\n", parentDoc.Ref.ID, inv.CustomerRef)
	} else if qboId != inv.CustomerRef {
		log.Printf("Parent doc %s has business.qboCustomerId=%s but invoice uses %s; skipping.\n",
			parentDoc.Ref.ID, qboId, inv.CustomerRef)
	}

	return nil
}

func (s *OAuthService) findParentWithQboId(ctx context.Context, qboId string) (bool, error) {
	if qboId == "" {
		return false, nil
	}
	// A normal query on "parents" for field "business.qboCustomerId" = qboId
	snap, err := s.firestore.Collection("parents").
		Where("business.qboCustomerId", "==", qboId).
		Limit(1).
		Documents(ctx).GetAll()
	if err != nil {
		return false, err
	}
	return len(snap) > 0, nil
}

// getGlobalTokens is an example of retrieving your QBO tokens from Firestore or environment.
func (s *OAuthService) getGlobalTokens(ctx context.Context) (*oauth2.Token, error) {
	// If tokens are stored at "intuit/globalTokens" under a sub-doc named "intuitoauth",
	// we need to parse that exact structure.
	docSnap, err := s.firestore.Collection("intuit").Doc("globalTokens").Get(ctx)
	if err != nil {
		return nil, err
	}

	// We'll define a small struct to capture the nested "intuitoauth" part.
	var data struct {
		Intuitoauth struct {
			AccessToken  string    `firestore:"accessToken"`
			RefreshToken string    `firestore:"refreshToken"`
			TokenType    string    `firestore:"tokenType"`
			Expiry       time.Time `firestore:"expiry"`
		} `firestore:"intuitoauth"`
	}

	if err := docSnap.DataTo(&data); err != nil {
		return nil, err
	}

	// Now pull fields out of data.Intuitoauth
	return &oauth2.Token{
		AccessToken:  data.Intuitoauth.AccessToken,
		RefreshToken: data.Intuitoauth.RefreshToken,
		TokenType:    data.Intuitoauth.TokenType,
		Expiry:       data.Intuitoauth.Expiry,
	}, nil
}

// refreshAccessTokenIfExpired is a helper from previous code that checks if token is expired,
// and if so, refreshes it using the refresh token.
func (s *OAuthService) refreshAccessTokenIfExpired(ctx context.Context, tok *oauth2.Token, oldRealmID string) (*oauth2.Token, error) {
	if tok.Valid() {
		return tok, nil
	}

	ts := s.config.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save newTok back under `intuitoauth`
	err = s.updateGlobalTokens(ctx, newTok, oldRealmID)
	if err != nil {
		return nil, err
	}

	return newTok, nil
}

func (s *OAuthService) updateGlobalTokens(ctx context.Context, newTok *oauth2.Token, realmID string) error {
	// Build a new TokenInfo.
	updated := TokenInfo{
		AccessToken:  newTok.AccessToken,
		RefreshToken: newTok.RefreshToken,
		TokenType:    newTok.TokenType,
		Expiry:       newTok.Expiry,
		RealmID:      realmID,
	}

	// Write it under "intuitoauth" just like storeTokens does.
	_, err := s.firestore.Collection("intuit").Doc("globalTokens").
		Set(ctx, map[string]interface{}{
			"intuitoauth": updated,
		}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update global tokens sub-document: %w", err)
	}
	return nil
}
