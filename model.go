package purchase

import (
	"context"
	"errors"
)

type PurchaseSystem interface {
	CreateOrder(ctx context.Context, no NewOrder) (OrderID, error)
	ReceivePayment(ctx context.Context, paymentEvet Payment) (PaymentID, error)
	UserHistory(ctx context.Context, userID UserID) ([]Payment, []Order, error)
	Pay(ctx context.Context, orderID OrderID, paymentIDs PaymentID) error
	UpdateTrackingInfo(ctx context.Context, orderID OrderID, update TrackingInfoUpdate) error
	ViewOrder(ctx context.Context, orderID OrderID) (Order, error)
	ViewReceipt(ctx context.Context, orderID OrderID) (Receipt, error)
}

var (
	ErrDoubleSpend           = errors.New("double spend")
	ErrInsufficientPayment   = errors.New("insufficent payment")
	ErrOrderNotFound         = errors.New("order not found")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrNetworkError          = errors.New("network error")
	ErrPaymentFound          = errors.New("payment found")
	ErrCreateOrderValidation = errors.New("invalid order")
)

type NewOrder struct {
	Fees      []Fee
	Products  []Product
	Currency  Currency
	Purchaser Purchaser
	Recipient Receipt
}

type Order struct {
	ID              OrderID
	Status          string
	UserID          UserID
	AmountRemaining int
	TotalAmount     int
}

type OrderID = string

type UserID = string

type PaymentID = string

type Payment struct {
	ID                   PaymentID
	GatewayID            string
	GatewayTransactionID string
	Amount               int
	Currency             Currency
	UserID               UserID
	Consumed             bool
	ConsumedBy           OrderID
}

type Product struct {
	ID        string
	Quantity  int
	UnitPrice int
}

type Currency = string

type Fee struct {
	Amount int
}

type Receipt struct{}

type Purchaser struct {
	ID UserID
}

type Recipient struct{}

type TrackingInfoUpdate struct{}

type CurrencyBalance struct{}

type NewOrderPolicy func(ctx context.Context, no NewOrder) (bool, error)

type PaymentRegistry interface {
	StoreOrder(ctx context.Context, order Order) error
	StorePayment(ctx context.Context, payment Payment) error
	FindPayment(ctx context.Context, ids PaymentID) (Payment, error)
	FindOrder(ctx context.Context, id OrderID) (Order, error)
	OrdersByUser(ctx context.Context, id UserID) ([]Order, error)
	PaymentsByUser(ctx context.Context, id UserID) ([]Payment, error)
}

const (
	OrderStatusPaidInFull     = "paid-in-full"
	OrderStatusPendingPayment = "pending-payment"
	OrderStatusPaidInPart     = "paid-in-part"
	OrderStatusOverPaid       = "over-paid"
)
