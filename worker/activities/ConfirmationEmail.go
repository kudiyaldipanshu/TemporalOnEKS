package activities

import (
	"context"
	"fmt"
	"temporal/worker/models"

	"go.temporal.io/sdk/activity"
)

type EmailResult struct {
	OrderID string
	Status  string
	Message string
}

func SendConfirmationEmail(ctx context.Context, input models.OrderInput) (EmailResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SendConfirmationEmail started", "orderID", input.OrderID)

	if err := checkEmailAlreadySent(input.OrderID); err != nil {
		logger.Info("SendConfirmationEmail: email already sent, skipping", "orderID", input.OrderID)
		return EmailResult{
			OrderID: input.OrderID,
			Status:  "SKIPPED",
			Message: "confirmation email already sent",
		}, nil
	}

	email := buildConfirmationEmail(input)

	if err := mockSendEmail(email); err != nil {
		logger.Error("SendConfirmationEmail: failed to send email",
			"orderID", input.OrderID,
			"error", err,
		)
		return EmailResult{}, err
	}

	markEmailSent(input.OrderID)

	fmt.Printf("[email] Confirmation sent for order %s (item: %s, qty: %d)\n",
		input.OrderID, input.ItemID, input.Qty)

	logger.Info("SendConfirmationEmail succeeded", "orderID", input.OrderID)
	return EmailResult{
		OrderID: input.OrderID,
		Status:  "COMPLETE",
		Message: fmt.Sprintf("confirmation email sent for order %s", input.OrderID),
	}, nil

}

type EmailPayload struct {
	To      string
	Subject string
	Body    string
}

func buildConfirmationEmail(input models.OrderInput) EmailPayload {
	return EmailPayload{
		To:      fmt.Sprintf("customer+%s@example.com", input.OrderID),
		Subject: fmt.Sprintf("Order Confirmation - %s", input.OrderID),
		Body: fmt.Sprintf(
			"Thank you for your order!\n\nOrder ID: %s\nItem: %s\nQuantity: %d\nStatus: Confirmed",
			input.OrderID, input.ItemID, input.Qty,
		),
	}
}

func mockSendEmail(email EmailPayload) error {
	// Swap this out for a real SES/SMTP/SendGrid call in production
	fmt.Printf("[mock-smtp] To: %s | Subject: %s\n", email.To, email.Subject)
	return nil
}

var sentEmails = map[string]bool{}

func checkEmailAlreadySent(orderID string) error {
	if sentEmails[orderID] {
		return fmt.Errorf("email already sent for order %s", orderID)
	}
	return nil
}

func markEmailSent(orderID string) {
	sentEmails[orderID] = true
}
