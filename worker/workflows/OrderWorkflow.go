package workflows

import (
	"temporal/worker/activities"
	"temporal/worker/models"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OrderWorkflow(ctx workflow.Context, input models.OrderInput) (activities.EmailResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", input.OrderID)

	validateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2,
		},
	})

	if err := workflow.ExecuteActivity(validateCtx, activities.ValidateOrder, input).Get(ctx, nil); err != nil {
		logger.Error("ValidateOrder failed", "orderID", input.OrderID, "error", err)
		return activities.EmailResult{}, err
	}

	logger.Info("ValidateOrder complete", "orderID", input.OrderID)

	reserveCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        5,
			InitialInterval:        3 * time.Second,
			BackoffCoefficient:     5.0,
			NonRetryableErrorTypes: []string{"ItemNotFound", "InsufficientStock"},
		},
	})

	if err := workflow.ExecuteActivity(reserveCtx, activities.ReserveInventory, input).Get(ctx, nil); err != nil {
		logger.Error("ReserveInventory failed", "orderID", input.OrderID, "error", err)
		return activities.EmailResult{}, err
	}
	logger.Info("ReserveInventory complete", "orderID", input.OrderID)

	chargeCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 10.0,
		},
	})

	if err := workflow.ExecuteActivity(chargeCtx, activities.ChargePayment, input).Get(ctx, nil); err != nil {
		logger.Error("ChargePayment failed", "orderID", input.OrderID, "error", err)
		return activities.EmailResult{}, err
	}
	logger.Info("ChargePayment complete", "orderID", input.OrderID)

	emailCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
		},
	})

	var result activities.EmailResult

	if err := workflow.ExecuteActivity(emailCtx, activities.SendConfirmationEmail, input).Get(ctx, &result); err != nil {
		logger.Error("SendConfirmationEmail failed", "orderID", input.OrderID, "error", err)
		return activities.EmailResult{}, err
	}
	logger.Info("SendConfirmationEmail complete", "orderID", input.OrderID)

	logger.Info("OrderWorkflow complete",
		"orderID", result.OrderID,
		"status", result.Status,
		"message", result.Message,
	)
	return result, nil

}
