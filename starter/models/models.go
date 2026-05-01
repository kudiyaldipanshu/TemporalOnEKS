package models

// type OrderInput struct {
// 	OrderID string
// 	ItemID  string
// 	Qty     int
// }

type OrderInput struct {
	OrderID string `json:"orderId"`
	ItemID  string `json:"itemId"`
	Qty     int    `json:"qty"`
}

// type OrderStatus string

// const (
// 	StatusCreated   OrderStatus = "CREATED"
// 	StatusValidated OrderStatus = "VALIDATED"
// 	StatusReserved  OrderStatus = "INVENTORY_RESERVED"
// 	StatusPaid      OrderStatus = "PAID"
// 	StatusCompleted OrderStatus = "COMPLETED"
// )

// type OrderState struct {
// 	OrderID string
// 	Status  OrderStatus
// }
