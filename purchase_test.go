package purchase_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/schafer14/purchase"
	"github.com/schafer14/purchase/purchase_registries/inmemory"
)

type contextKey int

const (
	ckPurchaseSystem contextKey = iota
	ckUserID
	ckPaymentID
	ckOrderID
)

var (
	orderIDs   []purchase.OrderID   = []purchase.OrderID{}
	paymentIDs []purchase.PaymentID = []purchase.PaymentID{}
	errors     []error              = []error{}
)

func ps(ctx context.Context) purchase.PurchaseSystem {
	return ctx.Value(ckPurchaseSystem).(purchase.PurchaseSystem)
}

func paymentUsed(ctx context.Context, paymentNum int, orderNum int) error {

	if err := ps(ctx).Pay(ctx, orderIDs[orderNum-1], paymentIDs[paymentNum-1]); err != nil {
		return err
	}

	return nil
}

func aPaymentIsConsumed(ctx context.Context, paymentNum int) error {
	userID := ctx.Value(ckUserID).(purchase.UserID)

	var payment purchase.Payment
	payments, _, err := ps(ctx).UserHistory(ctx, userID)
	if err != nil {
		return err
	}

	for _, p := range payments {
		if p.ID == paymentIDs[paymentNum-1] {
			payment = p
		}
	}

	if !payment.Consumed {
		return fmt.Errorf("payment is not marked as consumed")
	}

	return nil
}

func givenAnOrder(ctx context.Context, numOrders int, dollars, cents int, currency string) error {

	for i := 0; i < numOrders; i++ {
		newOrder := purchase.NewOrder{
			Products: []purchase.Product{
				{ID: "978-0-321-12521-7", Quantity: 1, UnitPrice: dollars*100 + cents},
			},
			Currency: purchase.Currency(currency),
			Purchaser: purchase.Purchaser{
				ID: ctx.Value(ckUserID).(purchase.UserID),
			},
		}

		orderID, err := ctx.Value(ckPurchaseSystem).(purchase.PurchaseSystem).CreateOrder(ctx, newOrder)
		if err != nil {
			return err
		}

		orderIDs = append(orderIDs, orderID)
	}

	return nil
}

func orderStatusOf(ctx context.Context, orderNum int, orderStatus string) error {
	order, err := ps(ctx).ViewOrder(ctx, orderIDs[orderNum-1])
	if err != nil {
		return err
	}

	if order.Status != orderStatus {
		return fmt.Errorf("order status should be '%s', got '%s'", orderStatus, order.Status)
	}

	return nil
}

func orderAmountRemaining(ctx context.Context, orderNum int, dollars, cents int, currency string) error {
	order, err := ps(ctx).ViewOrder(ctx, orderIDs[orderNum-1])
	if err != nil {
		return err
	}

	if order.AmountRemaining != dollars*100+cents {
		return fmt.Errorf("the amount remaining should be '%d', got '%d'", dollars*100+cents, order.AmountRemaining)
	}

	return nil

}

func paymentReceived(ctx context.Context, dollars, cents int, currency string) error {
	payment := purchase.Payment{
		GatewayID:            "visa",
		GatewayTransactionID: "xyz-123",
		Amount:               dollars*100 + cents,
		Currency:             currency,
		UserID:               ctx.Value(ckUserID).(purchase.UserID),
	}

	paymentID, err := ps(ctx).ReceivePayment(ctx, payment)
	if err != nil {
		return err
	}
	paymentIDs = append(paymentIDs, paymentID)

	return nil
}

func invalidOrder(ctx context.Context) error {

	newOrder := purchase.NewOrder{
		Products: []purchase.Product{
			{ID: "invalid-product-id", Quantity: 1, UnitPrice: 3},
		},
		Purchaser: purchase.Purchaser{
			ID: ctx.Value(ckUserID).(purchase.UserID),
		},
	}

	_, err := ctx.Value(ckPurchaseSystem).(purchase.PurchaseSystem).CreateOrder(ctx, newOrder)
	if err != nil {
		errors = append(errors, err)
	}

	return nil
}

func errorReturned(ctx context.Context, errorMsg string) error {
	if len(errors) != 1 {
		return fmt.Errorf("expected a single error")
	}

	if !strings.Contains(errors[0].Error(), errorMsg) {
		return fmt.Errorf(`expected error msg to contain "%s", got "%s"`, errorMsg, errors[0].Error())
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {

		registry := inmemory.New()
		purchaseSystem := purchase.New(registry, newOrderPolicy)

		ctx = context.WithValue(ctx, ckPurchaseSystem, purchaseSystem)
		ctx = context.WithValue(ctx, ckUserID, "user-1")
		orderIDs = []purchase.OrderID{}
		paymentIDs = []purchase.PaymentID{}
		errors = []error{}

		return ctx, nil
	})

	ctx.Step(`^payment (\d+) should be consumed$`, aPaymentIsConsumed)
	ctx.Step(`^(\d+) order with a total payment of \$(\d+)\.(\d+) ([A-Z]+)$`, givenAnOrder)
	ctx.Step(`^order (\d+) should be marked as ([A-Za-z\-]+)$`, orderStatusOf)
	ctx.Step(`^order (\d+) should have \$(-?\d+)\.(\d+) ([A-Z]+) remaining$`, orderAmountRemaining)
	ctx.Step(`^a payment is received for \$(\d+)\.(\d+) ([A-Z]+)$`, paymentReceived)
	ctx.Step(`^payment (\d+) is used for order (\d+)$`, paymentUsed)
	ctx.Step(`^an order fails to pass validation$`, invalidOrder)
	ctx.Step(`^a "([A-Za-z ]+)" error should be returned$`, errorReturned)

}

func newOrderPolicy(ctx context.Context, no purchase.NewOrder) (bool, error) {
	if no.Products[0].ID == "invalid-product-id" {
		return false, nil
	}
	return true, nil
}
