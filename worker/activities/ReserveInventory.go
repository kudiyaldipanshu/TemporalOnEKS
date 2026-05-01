package activities

import (
	"context"
	"fmt"
	"temporal/worker/models"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

var reservedOrders = map[string]bool{}

func ReserveInventory(ctx context.Context, input models.OrderInput) error {
	logger := activity.GetLogger(ctx)

	logger.Info("ReserveInventory started:", "orderID", input.OrderID, "itemID", input.ItemID, "qty", input.Qty)

	if reservedOrders[input.OrderID] {
		logger.Info("ReserveInventory: order already reserved, skipping", "orderID", input.OrderID)
		return nil
	}

	steps := 3

	for i := 1; i <= steps; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		activity.RecordHeartbeat(ctx, fmt.Sprintf("Reserving inventory step %d/%d", i, steps))

		time.Sleep(3 * time.Second)
	}

	stock, exists := mockInventory[input.ItemID]

	if !exists {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("item %s not found at reservation time", input.ItemID),
			"ItemNotFound",
			nil,
		)
	}

	if stock < input.Qty {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("stock depleted for item %s: need %d, have %d", input.ItemID, input.Qty, stock),
			"InsufficientStock",
			nil,
		)
	}

	mockInventory[input.ItemID] -= input.Qty
	reservedOrders[input.OrderID] = true

	logger.Info("ReserveInventory succeeded", "orderID", input.OrderID, "remainingStock", mockInventory[input.ItemID])
	return nil

}
