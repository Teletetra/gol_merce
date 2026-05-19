package service

import (
	"context"
	"time"

	"ecommerce_go/internal/domain"
)

type CheckoutService struct {
	carts    domain.CartRepository
	products domain.ProductRepository
	orders   domain.OrderRepository
	payments domain.PaymentRepository
	gateway  Gateway
}

func NewCheckoutService(
	carts domain.CartRepository,
	products domain.ProductRepository,
	orders domain.OrderRepository,
	payments domain.PaymentRepository,
	gateway Gateway,
) *CheckoutService {
	return &CheckoutService{
		carts:    carts,
		products: products,
		orders:   orders,
		payments: payments,
		gateway:  gateway,
	}
}

type CheckoutInput struct {
	ShippingAddress domain.Address `json:"shipping_address"`
	PaymentMethod   string         `json:"payment_method"`
	CardToken       string         `json:"card_token"`
	WalletID        string         `json:"wallet_id"`
	CustomerEmail   string         `json:"customer_email"`
}

func (s *CheckoutService) Checkout(ctx context.Context, userID string, input CheckoutInput) (domain.Order, domain.Payment, error) {
	cart, err := s.carts.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Order{}, domain.Payment{}, err
	}
	if len(cart.Items) == 0 {
		return domain.Order{}, domain.Payment{}, ErrValidation
	}

	orderItems := make([]domain.OrderItem, 0, len(cart.Items))
	var subtotal int64
	currency := "USD"
	reservedProducts := make([]domain.Product, 0, len(cart.Items))

	for _, cartItem := range cart.Items {
		product, found, err := s.products.FindByID(ctx, cartItem.ProductID)
		if err != nil {
			return domain.Order{}, domain.Payment{}, err
		}
		if !found || !product.Active || product.Stock < cartItem.Quantity {
			return domain.Order{}, domain.Payment{}, ErrValidation
		}

		product.Stock -= cartItem.Quantity
		if _, err := s.products.Update(ctx, product); err != nil {
			return domain.Order{}, domain.Payment{}, err
		}
		reservedProducts = append(reservedProducts, product)

		itemSubtotal := product.Price * int64(cartItem.Quantity)
		subtotal += itemSubtotal
		currency = product.Currency
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			UnitPrice:   product.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    itemSubtotal,
		})
	}

	tax := subtotal / 10
	shipping := int64(1500)
	var discount int64
	if subtotal > 20000 {
		discount = 1000
	}

	order := domain.Order{
		ID:              newID("ord"),
		UserID:          userID,
		Items:           orderItems,
		Subtotal:        subtotal,
		Tax:             tax,
		ShippingFee:     shipping,
		Discount:        discount,
		GrandTotal:      subtotal + tax + shipping - discount,
		Currency:        currency,
		Status:          domain.OrderStatusPending,
		ShippingAddress: input.ShippingAddress,
		CreatedAt:       time.Now().UTC(),
	}

	order, err = s.orders.Create(ctx, order)
	if err != nil {
		return domain.Order{}, domain.Payment{}, err
	}

	payment, chargeErr := s.gateway.Charge(ctx, ChargeInput{
		OrderID:       order.ID,
		Amount:        order.GrandTotal,
		Currency:      order.Currency,
		Method:        input.PaymentMethod,
		CardToken:     input.CardToken,
		WalletID:      input.WalletID,
		CustomerEmail: input.CustomerEmail,
	})

	payment, err = s.payments.Create(ctx, payment)
	if err != nil {
		return domain.Order{}, domain.Payment{}, err
	}

	order.PaymentID = payment.ID
	if chargeErr != nil {
		for _, product := range reservedProducts {
			for _, item := range cart.Items {
				if item.ProductID == product.ID {
					product.Stock += item.Quantity
					_, _ = s.products.Update(ctx, product)
				}
			}
		}
		order.Status = domain.OrderStatusCancelled
		if _, err := s.orders.Update(ctx, order); err != nil {
			return domain.Order{}, domain.Payment{}, err
		}
		return order, payment, chargeErr
	}

	order.Status = domain.OrderStatusPaid
	order, err = s.orders.Update(ctx, order)
	if err != nil {
		return domain.Order{}, domain.Payment{}, err
	}

	if err := s.carts.Clear(ctx, userID); err != nil {
		return domain.Order{}, domain.Payment{}, err
	}

	return order, payment, nil
}

func (s *CheckoutService) ListOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	return s.orders.FindByUserID(ctx, userID)
}
