package hours

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
)

type Service struct {
	firestore *firestore.Client
}

// NewService creates a new hours.Service.
func NewService(fs *firestore.Client) *Service {
	return &Service{firestore: fs}
}

// UpdateParentHours retrieves the parent's QBO Customer ID from the sub-document "business.qboCustomerId",
// then iterates over 'intuit/<qboCustomerId>/invoices' to sum HoursPurchased,
// and writes the total to "business.lifetime_hours" in the parent doc.
func (s *Service) UpdateParentHours(ctx context.Context, parentID string) error {
	// 1. Load parent doc
	parentRef := s.firestore.Collection("parents").Doc(parentID)
	parentSnap, err := parentRef.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get parent doc: %w", err)
	}

	parentData := parentSnap.Data()

	// 2. Read sub-document 'business' if present
	business, ok := parentData["business"].(map[string]interface{})
	if !ok {
		// If there's no business field, nothing to do
		log.Printf("Parent %s has no 'business' sub-document\n", parentID)
		return nil
	}

	qboCust, _ := business["qboCustomerId"].(string)
	if qboCust == "" {
		// no QBO customer ID => can't sum hours
		log.Printf("Parent %s has no qboCustomerId in business sub-doc\n", parentID)
		return nil
	}

	// 3. Iterate over all invoices in `intuit/<qboCustomerId>/invoices`
	invoicesRef := s.firestore.Collection("intuit").Doc(qboCust).Collection("invoices")
	invoicesSnap, err := invoicesRef.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get invoices for customerRef=%s: %w", qboCust, err)
	}

	var totalHours float64
	for _, invDoc := range invoicesSnap {
		invData := invDoc.Data()
		// Double-check that invData["customerRef"] == qboCust, just for safety
		if custRef, ok2 := invData["customerRef"].(string); ok2 {
			if custRef != qboCust {
				// mismatch scenario, skip or log
				log.Printf("Warning: invoice %s has customerRef=%s not matching qbo=%s\n", invDoc.Ref.ID, custRef, qboCust)
				continue
			}
		}
		// read "hoursPurchased" if present
		hp, _ := invData["hoursPurchased"].(float64)
		totalHours += hp
	}

	log.Printf("Calculated total hours purchased for parent %s (qbo=%s): %.2f\n", parentID, qboCust, totalHours)

	// 4. Write totalHours to parent's sub-document field 'business.lifetime_hours'
	// We'll update the 'business' map in place, or do a direct doc update.
	business["lifetime_hours"] = totalHours

	// MergeAll to keep existing fields
	_, err = parentRef.Set(ctx, map[string]interface{}{
		"business": business,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update parent doc with lifetime_hours: %w", err)
	}

	log.Printf("Parent %s updated business.lifetime_hours=%.2f\n", parentID, totalHours)
	return nil
}
