package activities

import (
	"context"
	"fmt"
	"math/rand"
	"temporal/worker/models"

	"go.temporal.io/sdk/activity"
)

func ChargePayment(ctx context.Context, input models.OrderInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("ChargePayment started", "orderID", input.OrderID)

	attemptInfo := activity.GetInfo(ctx)

	logger.Info("ChargePayment attempt", "attempt", attemptInfo.Attempt)

	if err := mockPaymentGateway(input.OrderID); err != nil {

		logger.Warn("ChargePayment: payment gateway failed, will retry",
			"orderID", input.OrderID,
			"attempt", attemptInfo.Attempt,
			"error", err,
		)
		return err // retryable — Temporal will retry up to max attempts
	}

	logger.Info("ChargePayment succeeded", "orderID", input.OrderID)
	return nil

}

func mockPaymentGateway(orderID string) error {
	if rand.Intn(2) == 0 {
		return fmt.Errorf("payment gateway error: transaction declined for order %s", orderID)
	}
	return nil

}
