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

func (s *OAuthService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var webhook WebhookEvent
	var lastCustomerRef string
	if err := json.Unmarshal(bodyBytes, &webhook); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	for _, note := range webhook.EventNotifications {
		realmID := note.RealmID

		for _, entity := range note.DataChangeEvent.Entities {

			log.Printf("[Webhook Debug] entity.Name=%s, entity.ID=%s, entity.Operation=%s",
				entity.Name, entity.ID, entity.Operation)
			// handle invoices
			if strings.EqualFold(entity.Name, "Invoice") {
				invoiceID := entity.ID
				operation := entity.Operation

				if isDeleteOperation(operation) {
					cRef, err := s.deleteInvoiceDoc(r.Context(), invoiceID)
					if err != nil {
						log.Printf("Failed to delete invoice doc %s: %v\n", invoiceID, err)
					} else {
						log.Printf("Webhook: successfully handled delete event for invoice %s\n", invoiceID)
						if cRef != "" {
							lastCustomerRef = cRef
						}
					}
					continue
				}

				// otherwise handle create/update/void
				invData, err := s.fetchInvoiceFromQBO(r.Context(), realmID, invoiceID)
				if err != nil {
					log.Printf("Failed to fetch invoice %s from realm %s: %v\n", invoiceID, realmID, err)
					continue
				}
				lastCustomerRef = invData.CustomerRef
				// Store it in Firestore: intuit/<customerRef>/invoices/<invoiceID>
				err = s.storeInvoice(r.Context(), invData)
				if err != nil {
					log.Printf("Failed to store invoice in Firestore: %v\n", err)
				} else {
					log.Printf("Webhook: successfully handled %s event for invoice %s\n", operation, invoiceID)
				}
			}

			// handle payments
			if strings.EqualFold(entity.Name, "Payment") {
				paymentID := entity.ID
				operation := entity.Operation

				if isDeleteOperation(operation) {
					err := s.deletePaymentDoc(r.Context(), paymentID)
					if err != nil {
						log.Printf("Failed to delete payment doc %s: %v", paymentID, err)
					} else {
						log.Printf("Deleted payment doc for PaymentID=%s", paymentID)
					}
					continue
				}

				payData, err := s.fetchPaymentFromQBO(r.Context(), realmID, paymentID)
				if err != nil {
					log.Printf("Failed to fetch payment %s: %v", paymentID, err)
					continue
				}

				if len(payData.Lines) > 0 && payData.Lines[0].InvoiceID != "" {
					invSnap, _ := s.findInvoiceDocByID(r.Context(), payData.Lines[0].InvoiceID)
					if invSnap != nil {
						var inv InvoiceRecord
						if e := invSnap.DataTo(&inv); e == nil {
							lastCustomerRef = inv.CustomerRef
						}
					}
				}

				if err := s.storePayment(r.Context(), payData); err != nil {
					log.Printf("Failed to store payment doc: %v", err)
				} else {
					log.Printf("Stored payment doc for PaymentID=%s", paymentID)
				}
			}

			// handle credit payments
			if strings.EqualFold(entity.Name, "CreditMemo") {
				creditMemoID := entity.ID
				operation := entity.Operation

				if isDeleteOperation(operation) {
					err := s.deleteCreditMemoDoc(r.Context(), creditMemoID)
					if err != nil {
						log.Printf("Failed to delete credit memo doc %s: %v", creditMemoID, err)
					} else {
						log.Printf("Deleted credit memo doc for creditMemoID=%s", creditMemoID)
					}
					continue
				}

				cmData, err := s.fetchCreditMemoFromQBO(r.Context(), realmID, creditMemoID)
				if err != nil {
					log.Printf("Failed to fetch creditMemo %s: %v", creditMemoID, err)
					continue
				}

				if len(cmData.Lines) > 0 && cmData.Lines[0].InvoiceID != "" {
					invSnap, _ := s.findInvoiceDocByID(r.Context(), cmData.Lines[0].InvoiceID)
					if invSnap != nil {
						var inv InvoiceRecord
						if e := invSnap.DataTo(&inv); e == nil {
							lastCustomerRef = inv.CustomerRef
						}
					}
				}

				if err := s.storeCreditMemo(r.Context(), cmData); err != nil {
					log.Printf("Failed to store credit memo doc: %v", err)
				} else {
					log.Printf("Stored credit memo doc for creditMemoID=%s", creditMemoID)
				}
			}

		}
	}

	if lastCustomerRef != "" {
		err := s.recalcTotalBalance(r.Context(), lastCustomerRef)
		if err != nil {
			log.Printf("Failed to recalc total_balance for customerRef=%s: %v", lastCustomerRef, err)
		}
	}

	// Respond 200 to acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// GENERAL HELPER
func isDeleteOperation(op string) bool {
	return strings.EqualFold(op, "Delete") ||
		strings.EqualFold(op, "Remove") ||
		strings.EqualFold(op, "Deleted") ||
		strings.EqualFold(op, "Removed")
}

func (s *OAuthService) recalcTotalBalance(ctx context.Context, customerRef string) error {
	parentDocRef := s.firestore.Collection("intuit").Doc(customerRef)
	invoiceColRef := parentDocRef.Collection("invoices")

	snaps, err := invoiceColRef.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to fetch invoices for %s: %w", customerRef, err)
	}

	var totalBalance float64
	var totalHours float64

	for _, snap := range snaps {
		var inv InvoiceRecord
		if err := snap.DataTo(&inv); err == nil {
			totalBalance += inv.Balance
			totalHours += inv.HoursPurchased
		} else {
			log.Printf("Warning: couldn't parse invoice doc %s for customer %s: %v",
				snap.Ref.ID, customerRef, err)
		}
	}

	// Update both 'total_balance' and 'total_hours' in the parent doc
	_, err = parentDocRef.Set(ctx, map[string]interface{}{
		"total_balance": totalBalance,
		"total_hours":   totalHours,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update totals for %s: %w", customerRef, err)
	}

	log.Printf("Recalculated total_balance=%.2f and total_hours=%.2f for customerRef=%s",
		totalBalance, totalHours, customerRef)
	return nil
}

// HANDLE CREDIT PAYMENTS VIA CREDIT MEMO AND STRUCTS BELOW
type CreditMemoRecord struct {
	CreditMemoID string           `firestore:"creditMemoID,omitempty"`
	DocNumber    string           `firestore:"docNumber,omitempty"`
	TotalAmt     float64          `firestore:"totalAmt,omitempty"`
	TxnDate      time.Time        `firestore:"txnDate,omitempty"`
	Lines        []CreditMemoLine `firestore:"lines,omitempty"`
}

type CreditMemoLine struct {
	InvoiceID string  `firestore:"invoiceID,omitempty"`
	Amount    float64 `firestore:"amount,omitempty"`
}

func (s *OAuthService) fetchCreditMemoFromQBO(ctx context.Context, realmID, creditMemoID string) (*CreditMemoRecord, error) {
	token, err := s.getGlobalTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}
	validToken, err := s.refreshAccessTokenIfExpired(ctx, token, realmID)
	if err != nil {
		return nil, fmt.Errorf("refresh token error: %w", err)
	}

	client := s.config.Client(ctx, validToken)
	url := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/creditmemo/%s?minorversion=65", realmID, creditMemoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create creditmemo request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creditmemo fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("creditmemo fetch returned status %d", resp.StatusCode)
	}

	var cmResp struct {
		CreditMemo struct {
			ID        string  `json:"Id"`
			DocNumber string  `json:"DocNumber"`
			TotalAmt  float64 `json:"TotalAmt"`
			TxnDate   string  `json:"TxnDate"`
			Line      []struct {
				Amount    float64 `json:"Amount"`
				LinkedTxn []struct {
					TxnId   string `json:"TxnId"`
					TxnType string `json:"TxnType"`
				} `json:"LinkedTxn"`
			} `json:"Line"`
		} `json:"CreditMemo"`
		Time string `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cmResp); err != nil {
		return nil, fmt.Errorf("failed to decode creditmemo JSON: %w", err)
	}

	c := cmResp.CreditMemo
	cmRecord := &CreditMemoRecord{
		CreditMemoID: c.ID,
		DocNumber:    c.DocNumber,
		TotalAmt:     c.TotalAmt,
	}

	// parse TxnDate
	if c.TxnDate != "" {
		if dt, err2 := time.Parse("2006-01-02", c.TxnDate); err2 == nil {
			cmRecord.TxnDate = dt
		}
	}

	// parse lines referencing Invoices
	var lines []CreditMemoLine
	for _, l := range c.Line {
		for _, linked := range l.LinkedTxn {
			if strings.EqualFold(linked.TxnType, "Invoice") {
				lines = append(lines, CreditMemoLine{
					InvoiceID: linked.TxnId,
					Amount:    l.Amount,
				})
			}
		}
	}
	cmRecord.Lines = lines

	return cmRecord, nil
}

func (s *OAuthService) storeCreditMemo(ctx context.Context, cm *CreditMemoRecord) error {
	if cm == nil || cm.CreditMemoID == "" {
		return fmt.Errorf("missing CreditMemoRecord or CreditMemoID")
	}

	for _, line := range cm.Lines {
		if line.InvoiceID == "" || line.Amount <= 0 {
			continue
		}

		invoiceDocSnap, err := s.findInvoiceDocByID(ctx, line.InvoiceID)
		if err != nil {
			log.Printf("Could not find invoice doc for invoiceID=%s: %v", line.InvoiceID, err)
			continue
		}
		if invoiceDocSnap == nil {
			continue
		}

		invoiceDocRef := invoiceDocSnap.Ref

		credRef := invoiceDocRef.Collection("credits").Doc(cm.CreditMemoID)

		data := map[string]interface{}{
			"creditMemoID":    cm.CreditMemoID,
			"docNumber":       cm.DocNumber,
			"txnDate":         cm.TxnDate,
			"amountApplied":   line.Amount,
			"total_creditAmt": cm.TotalAmt,
			"created_at":      time.Now(),
		}

		if _, err := credRef.Set(ctx, data); err != nil {
			log.Printf("Failed to create credit memo doc in subcollection: %v", err)
			continue
		}

		// decrement invoice’s balance
		_, err = invoiceDocRef.Update(ctx, []firestore.Update{
			{Path: "balance", Value: firestore.Increment(-line.Amount)},
		})
		if err != nil {
			log.Printf("Failed to update invoice balance for invoiceID=%s: %v", line.InvoiceID, err)
		} else {
			log.Printf("Updated invoice %s balance by subtracting credit %.2f", line.InvoiceID, line.Amount)
		}
	}
	return nil
}

func (s *OAuthService) deleteCreditMemoDoc(ctx context.Context, creditMemoID string) error {
	colRef := s.firestore.CollectionGroup("credits").
		Where("creditMemoID", "==", creditMemoID)

	snaps, err := colRef.Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	for _, snap := range snaps {

		_, err := snap.Ref.Delete(ctx)
		if err != nil {
			log.Printf("Failed to delete credit memo doc %s in invoice subcollection: %v", creditMemoID, err)
		}
	}
	return nil
}

// HANDLE PAYMENT METHODS AND STRUCTS BELOW
type PaymentRecord struct {
	PaymentID     string        `firestore:"paymentID,omitempty"`
	PaymentRefNum string        `firestore:"paymentRefNum,omitempty"` // e.g. check # or reference
	PaymentMethod string        `firestore:"paymentMethod,omitempty"` // e.g. “Visa”, “Check”
	TotalAmt      float64       `firestore:"totalAmt,omitempty"`
	TxnDate       time.Time     `firestore:"txnDate,omitempty"`
	Lines         []PaymentLine `firestore:"lines,omitempty"` // each line references an invoice
}

type PaymentLine struct {
	InvoiceID string  `firestore:"invoiceID,omitempty"` // QBO internal invoice Id
	Amount    float64 `firestore:"amount,omitempty"`    // portion allocated to that invoice
}

func (s *OAuthService) fetchPaymentFromQBO(ctx context.Context, realmID, paymentID string) (*PaymentRecord, error) {
	token, err := s.getGlobalTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}
	validToken, err := s.refreshAccessTokenIfExpired(ctx, token, realmID)
	if err != nil {
		return nil, fmt.Errorf("refresh token error: %w", err)
	}

	client := s.config.Client(ctx, validToken)
	url := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/payment/%s?minorversion=65", realmID, paymentID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payment fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment fetch returned status %d", resp.StatusCode)
	}

	var payResp struct {
		Payment struct {
			ID               string  `json:"Id"`
			PaymentRefNum    string  `json:"PaymentRefNum"` // e.g. check # or reference
			TotalAmt         float64 `json:"TotalAmt"`
			TxnDate          string  `json:"TxnDate"` // e.g. 2025-02-04
			PaymentMethodRef struct {
				Name  string `json:"name,omitempty"` // sometimes QBO returns just a value
				Value string `json:"value"`
			} `json:"PaymentMethodRef"`

			Line []struct {
				Amount    float64 `json:"Amount"`
				LinkedTxn []struct {
					TxnId   string `json:"TxnId"`
					TxnType string `json:"TxnType"`
				} `json:"LinkedTxn"`
			} `json:"Line"`
		} `json:"Payment"`
		Time string `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment JSON: %w", err)
	}

	p := payResp.Payment
	payment := &PaymentRecord{
		PaymentID:     p.ID,
		PaymentRefNum: p.PaymentRefNum,
		TotalAmt:      p.TotalAmt,
	}

	if p.TxnDate != "" {
		if dt, err2 := time.Parse("2006-01-02", p.TxnDate); err2 == nil {
			payment.TxnDate = dt
		}
	}

	if p.PaymentMethodRef.Name != "" {
		payment.PaymentMethod = p.PaymentMethodRef.Name
	} else {
		payment.PaymentMethod = p.PaymentMethodRef.Value
	}

	var lines []PaymentLine
	for _, lineItem := range p.Line {
		for _, linked := range lineItem.LinkedTxn {
			if strings.EqualFold(linked.TxnType, "Invoice") {
				lines = append(lines, PaymentLine{
					InvoiceID: linked.TxnId,
					Amount:    lineItem.Amount,
				})
			}
		}
	}
	payment.Lines = lines

	return payment, nil
}

func (s *OAuthService) storePayment(ctx context.Context, pay *PaymentRecord) error {
	if pay == nil || pay.PaymentID == "" {
		return fmt.Errorf("missing PaymentRecord or PaymentID")
	}

	//   intuit/<customerRef>/invoices/<invoiceID>/payments/<paymentID>
	for _, line := range pay.Lines {
		if line.InvoiceID == "" || line.Amount <= 0 {
			continue
		}

		invoiceDocSnap, err := s.findInvoiceDocByID(ctx, line.InvoiceID)
		if err != nil {
			log.Printf("Could not find invoice doc for invoiceID=%s: %v", line.InvoiceID, err)
			continue
		}
		if invoiceDocSnap == nil {
			// no invoice doc found
			continue
		}

		invoiceDocRef := invoiceDocSnap.Ref

		paySubRef := invoiceDocRef.Collection("payments").Doc(pay.PaymentID)

		payDoc := map[string]interface{}{
			"paymentID":            pay.PaymentID,
			"paymentRefNum":        pay.PaymentRefNum,
			"paymentMethod":        pay.PaymentMethod,
			"txnDate":              pay.TxnDate,
			"payment_on_invoice":   line.Amount, // partial allocated amount to THIS invoice
			"total_payment_amount": pay.TotalAmt,
			"created_at":           time.Now(),
		}

		if _, err := paySubRef.Set(ctx, payDoc); err != nil {
			log.Printf("Failed to create payment doc in subcollection: %v", err)
			continue
		}

		_, err = invoiceDocRef.Update(ctx, []firestore.Update{
			{Path: "balance", Value: firestore.Increment(-line.Amount)},
		})
		if err != nil {
			log.Printf("Failed to update invoice balance for invoiceID=%s: %v", line.InvoiceID, err)
		} else {
			log.Printf("Updated invoice %s balance by subtracting %.2f", line.InvoiceID, line.Amount)
		}
	}

	return nil
}

func (s *OAuthService) deletePaymentDoc(ctx context.Context, paymentID string) error {
	colRef := s.firestore.CollectionGroup("payments").
		Where("paymentID", "==", paymentID)

	snaps, err := colRef.Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	for _, snap := range snaps {
		_, err := snap.Ref.Delete(ctx)
		if err != nil {
			log.Printf("Failed to delete Payment doc %s in invoice subcollection: %v", paymentID, err)
		}
	}
	return nil
}

func (s *OAuthService) findInvoiceDocByID(ctx context.Context, invoiceID string) (*firestore.DocumentSnapshot, error) {
	colRef := s.firestore.CollectionGroup("invoices").
		Where("invoiceID", "==", invoiceID).
		Limit(1)

	snaps, err := colRef.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, nil // not found
	}
	return snaps[0], nil
}

// HANDLE INVOICE METHODS AND STRUCTS BELOW
//
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

func (s *OAuthService) deleteInvoiceDoc(ctx context.Context, invoiceID string) (string, error) {
	log.Printf("deleteInvoiceDoc called for invoiceID=%s", invoiceID)
	colRef := s.firestore.CollectionGroup("invoices").
		Where("invoiceID", "==", invoiceID).
		Limit(1)

	snaps, err := colRef.Documents(ctx).GetAll()
	if err != nil {
		return "", err
	}
	if len(snaps) == 0 {
		return "", nil
	}

	var inv InvoiceRecord
	if err := snaps[0].DataTo(&inv); err != nil {
		log.Printf("Failed to parse invoice data for doc %s: %v", snaps[0].Ref.ID, err)
		// but we can still delete it
	}

	docRef := snaps[0].Ref
	if _, err := docRef.Delete(ctx); err != nil {
		return "", fmt.Errorf("failed to delete invoice doc %s: %w", invoiceID, err)
	}
	return inv.CustomerRef, nil
}

func (s *OAuthService) fetchCustomerFromQBO(ctx context.Context, realmID, customerID string) (string, error) {
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

	var custResp struct {
		Customer struct {
			Id               string `json:"Id"`
			PrimaryEmailAddr string `json:"PrimaryEmailAddr"`
		} `json:"Customer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&custResp); err != nil {
		return "", fmt.Errorf("failed to decode customer JSON: %w", err)
	}

	return custResp.Customer.PrimaryEmailAddr, nil
}

func (s *OAuthService) fetchInvoiceFromQBO(ctx context.Context, realmID, invoiceID string) (*InvoiceRecord, error) {

	token, err := s.getGlobalTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global tokens: %w", err)
	}

	validToken, err := s.refreshAccessTokenIfExpired(ctx, token, realmID)
	if err != nil {
		return nil, fmt.Errorf("refresh token error: %w", err)
	}

	client := s.config.Client(ctx, validToken)
	url := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/invoice/%s?minorversion=65", realmID, invoiceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build invoice fetch request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("invoice fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invoice fetch returned status %d", resp.StatusCode)
	}

	var invoiceResp struct {
		Invoice struct {
			ID          string `json:"Id"`
			DocNumber   string `json:"DocNumber"`
			CustomerRef struct {
				Value string `json:"value"`
			} `json:"CustomerRef"`
			BillEmail struct {
				Address string `json:"Address"`
			} `json:"BillEmail"`
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
				} `json:"SalesItemLineDetail,omitempty"`
			} `json:"Line"`
		} `json:"Invoice"`
		Time string `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to decode invoice JSON: %w", err)
	}

	invoice := invoiceResp.Invoice
	docNumber := invoice.DocNumber
	createTime, _ := time.Parse(time.RFC3339, invoice.MetaData.CreateTime)
	updateTime, _ := time.Parse(time.RFC3339, invoice.MetaData.LastUpdatedTime)

	billEmail := invoice.BillEmail.Address

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

	if err := s.autoAssociateParent(ctx, inv); err != nil {
		log.Printf("AutoAssociateParent fucntion error: %v\n", err)
	}

	return nil
}

func (s *OAuthService) autoAssociateParent(ctx context.Context, inv *InvoiceRecord) error {
	if inv.BillEmail == "" {
		return nil
	}

	alreadyLinked, err := s.findParentWithQboId(ctx, inv.CustomerRef)
	if err != nil {
		return err
	}
	if alreadyLinked {
		log.Printf("Skipping re-assign: qboCustomerId=%s is already linked\n", inv.CustomerRef)
		return nil
	}

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

	if qboId == "" {
		businessData["qboCustomerId"] = inv.CustomerRef
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
	snap, err := s.firestore.Collection("parents").
		Where("business.qboCustomerId", "==", qboId).
		Limit(1).
		Documents(ctx).GetAll()
	if err != nil {
		return false, err
	}
	return len(snap) > 0, nil
}

func (s *OAuthService) getGlobalTokens(ctx context.Context) (*oauth2.Token, error) {
	docSnap, err := s.firestore.Collection("intuit").Doc("globalTokens").Get(ctx)
	if err != nil {
		return nil, err
	}
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

	return &oauth2.Token{
		AccessToken:  data.Intuitoauth.AccessToken,
		RefreshToken: data.Intuitoauth.RefreshToken,
		TokenType:    data.Intuitoauth.TokenType,
		Expiry:       data.Intuitoauth.Expiry,
	}, nil
}

func (s *OAuthService) refreshAccessTokenIfExpired(ctx context.Context, tok *oauth2.Token, oldRealmID string) (*oauth2.Token, error) {
	if tok.Valid() {
		return tok, nil
	}

	ts := s.config.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	err = s.updateGlobalTokens(ctx, newTok, oldRealmID)
	if err != nil {
		return nil, err
	}

	return newTok, nil
}

func (s *OAuthService) updateGlobalTokens(ctx context.Context, newTok *oauth2.Token, realmID string) error {
	updated := TokenInfo{
		AccessToken:  newTok.AccessToken,
		RefreshToken: newTok.RefreshToken,
		TokenType:    newTok.TokenType,
		Expiry:       newTok.Expiry,
		RealmID:      realmID,
	}

	_, err := s.firestore.Collection("intuit").Doc("globalTokens").
		Set(ctx, map[string]interface{}{
			"intuitoauth": updated,
		}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update global tokens sub-document: %w", err)
	}
	return nil
}
