package purchase

import (
	"context"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

type service struct {
	registry       PaymentRegistry
	newOrderPolicy NewOrderPolicy
}

func New(registry PaymentRegistry, newOrderPolicy NewOrderPolicy) PurchaseSystem {
	return &service{registry, newOrderPolicy}
}

func (s *service) CreateOrder(ctx context.Context, no NewOrder) (OrderID, error) {
	passed, err := s.newOrderPolicy(ctx, no)
	if err != nil {
		return "", errors.Wrap(err, "create order: could not fetch newOrderPolicy")
	}

	if !passed {
		return "", ErrCreateOrderValidation
	}

	id := ksuid.New().String()

	total := s.calculateTotal(no)

	order := Order{
		ID:              id,
		Status:          OrderStatusPendingPayment,
		UserID:          no.Purchaser.ID,
		TotalAmount:     total,
		AmountRemaining: total,
	}
	if err := s.registry.StoreOrder(ctx, order); err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) ReceivePayment(ctx context.Context, paymentEvet Payment) (PaymentID, error) {
	id := ksuid.New().String()

	paymentEvet.ID = id
	if err := s.registry.StorePayment(ctx, paymentEvet); err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) UserHistory(ctx context.Context, userID UserID) ([]Payment, []Order, error) {
	orders, err := s.registry.OrdersByUser(ctx, userID)
	if err != nil {
		return []Payment{}, []Order{}, err
	}
	payments, err := s.registry.PaymentsByUser(ctx, userID)
	if err != nil {
		return []Payment{}, []Order{}, err
	}

	return payments, orders, nil
}

func (s *service) Pay(ctx context.Context, orderID OrderID, paymentID PaymentID) error {
	payment, err := s.registry.FindPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	order, err := s.registry.FindOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if payment.Consumed {
		return nil
	}

	payment.Consumed = true
	payment.ConsumedBy = order.ID

	if order.AmountRemaining > payment.Amount {
		order.Status = OrderStatusPaidInPart
	} else if order.AmountRemaining == payment.Amount {
		order.Status = OrderStatusPaidInFull
	} else {
		order.Status = OrderStatusOverPaid
	}
	order.AmountRemaining = order.AmountRemaining - payment.Amount

	if err := s.registry.StorePayment(ctx, payment); err != nil {
		return err
	}

	if err := s.registry.StoreOrder(ctx, order); err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateTrackingInfo(ctx context.Context, orderID OrderID, update TrackingInfoUpdate) error {
	return nil
}

func (s *service) ViewOrder(ctx context.Context, orderID OrderID) (Order, error) {
	return s.registry.FindOrder(ctx, orderID)
}

func (s *service) ViewReceipt(ctx context.Context, orderID OrderID) (Receipt, error) {
	return Receipt{}, nil
}

func (s *service) calculateTotal(no NewOrder) int {
	total := 0

	for _, p := range no.Products {
		total += p.UnitPrice * p.Quantity
	}

	for _, f := range no.Fees {
		total += f.Amount
	}

	return total
}
