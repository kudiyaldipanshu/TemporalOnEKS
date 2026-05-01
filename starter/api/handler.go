package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"temporal/starter/models"
	"time"

	"go.temporal.io/sdk/client"
)

type Handler struct {
	TemporalClient client.Client
}

func NewHandler(c client.Client) *Handler {
	return &Handler{TemporalClient: c}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log.Println("event=request_received path=/order method=", r.Method)

	var order models.OrderInput
	err := json.NewDecoder(r.Body).Decode(&order)

	if err != nil {
		log.Println("event=invalid_body error=", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	// Validate input
	if order.OrderID == "" || order.ItemID == "" || order.Qty <= 0 {
		log.Printf("event=validation_failed orderId=%s itemId=%s qty=%d",
			order.OrderID, order.ItemID, order.Qty)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	log.Printf("event=order_request_validated orderId=%s itemId=%s qty=%d",
		order.OrderID, order.ItemID, order.Qty)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	options := client.StartWorkflowOptions{
		ID:        "order-" + order.OrderID,
		TaskQueue: "order-queue",
	}

	// Start workflow
	we, err := h.TemporalClient.ExecuteWorkflow(
		ctx,
		options,
		"OrderWorkflow",
		order,
	)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Printf("event=workflow_start_failed orderId=%s error=%v",
			order.OrderID, err)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("event=workflow_started orderId=%s workflowId=%s",
		order.OrderID, we.GetID())

	response := map[string]string{
		"workflowId": we.GetID(),
		"orderId":    order.OrderID,
	}

	json.NewEncoder(w).Encode(response)
}
