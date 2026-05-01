package activities

import (
	"context"
	"fmt"
	"temporal/worker/models"

	"go.temporal.io/sdk/activity"
)

// Mock inventory — in production this would be a DB/cache lookup
var mockInventory = map[string]int{
	"ITEM-001": 100,
	"ITEM-002": 50,
	"ITEM-003": 0, // intentionally out of stock
}

func ValidateOrder(ctx context.Context, input models.OrderInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateOrder started", "orderID", input.OrderID, "itemID", input.ItemID, "qty", input.Qty)

	// Check 1: item must exist in catalog
	stock, exists := mockInventory[input.ItemID]
	if !exists {
		return fmt.Errorf("item %s does not exist in catalog", input.ItemID)
	}

	// Check 2: item must be in stock
	if stock == 0 {
		return fmt.Errorf("item %s is out of stock", input.ItemID)
	}

	// Check 3: requested qty must be fulfillable
	if input.Qty > stock {
		return fmt.Errorf("insufficient stock for item %s: requested %d, available %d", input.ItemID, input.Qty, stock)
	}

	// Check 4: qty must be positive
	if input.Qty <= 0 {
		return fmt.Errorf("invalid quantity %d: must be greater than zero", input.Qty)
	}

	logger.Info("ValidateOrder succeeded", "orderID", input.OrderID, "itemID", input.ItemID, "availableStock", stock)
	return nil
}
