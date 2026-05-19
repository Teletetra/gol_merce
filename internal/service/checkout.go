package service

import (
	"context"
	"strings"
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
	BillingAddress  domain.Address `json:"billing_address"`
	ShippingMethod  string         `json:"shipping_method"`
	CouponCode      string         `json:"coupon_code"`
	Notes           string         `json:"notes"`
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

	subtotal, currency, err := s.calculateCartSnapshot(ctx, cart.Items)
	if err != nil {
		return domain.Order{}, domain.Payment{}, err
	}

	orderItems := make([]domain.OrderItem, 0, len(cart.Items))
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
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			ProductSlug: product.Slug,
			SKU:         product.SKU,
			ImageURL:    primaryImage(product.Images),
			UnitPrice:   product.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    itemSubtotal,
			Attributes:  product.Attributes,
		})
	}
	cart.CouponCode = normalizeCouponCode(input.CouponCode, cart.CouponCode)
	if cart.CouponCode != "" {
		cart.Subtotal = subtotal
		cart.Tax = subtotal / 10
		cart.ShippingFee = shippingFeeByMethod(input.ShippingMethod, subtotal)
		cart.Discount = discountForCoupon(cart.CouponCode, subtotal)
		cart.GrandTotal = cart.Subtotal + cart.Tax + cart.ShippingFee - cart.Discount
		cart.Currency = currency
		cart, err = s.carts.Save(ctx, cart)
		if err != nil {
			return domain.Order{}, domain.Payment{}, err
		}
	}

	now := time.Now().UTC()
	order := domain.Order{
		ID:              newID("ord"),
		OrderNumber:     newOrderNumber(),
		UserID:          userID,
		Items:           orderItems,
		Subtotal:        subtotal,
		Tax:             subtotal / 10,
		ShippingFee:     shippingFeeByMethod(input.ShippingMethod, subtotal),
		Discount:        discountForCoupon(cart.CouponCode, subtotal),
		GrandTotal:      subtotal + (subtotal / 10) + shippingFeeByMethod(input.ShippingMethod, subtotal) - discountForCoupon(cart.CouponCode, subtotal),
		Currency:        currency,
		Status:          domain.OrderStatusPending,
		CouponCode:      cart.CouponCode,
		ShippingMethod:  defaultShippingMethod(input.ShippingMethod),
		Notes:           input.Notes,
		ShippingAddress: input.ShippingAddress,
		BillingAddress:  coalesceAddress(input.BillingAddress, input.ShippingAddress),
		Events: []domain.OrderEvent{
			{
				Type:      "order_created",
				Message:   "Order created and awaiting payment capture.",
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
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
		order.UpdatedAt = time.Now().UTC()
		order.Events = append(order.Events, domain.OrderEvent{
			Type:      "payment_failed",
			Message:   "Payment failed and reserved stock was restored.",
			CreatedAt: order.UpdatedAt,
		})
		if _, err := s.orders.Update(ctx, order); err != nil {
			return domain.Order{}, domain.Payment{}, err
		}
		return order, payment, chargeErr
	}

	paidAt := time.Now().UTC()
	order.Status = domain.OrderStatusPaid
	order.PaidAt = &paidAt
	order.UpdatedAt = paidAt
	order.Events = append(order.Events, domain.OrderEvent{
		Type:      "payment_captured",
		Message:   "Payment was captured successfully.",
		CreatedAt: paidAt,
	})
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

func defaultShippingMethod(method string) string {
	if method == "" {
		return "standard"
	}
	return method
}

func shippingFeeForSubtotal(subtotal int64) int64 {
	if subtotal >= 50000 {
		return 0
	}
	return 1500
}

func shippingFeeByMethod(method string, subtotal int64) int64 {
	switch defaultShippingMethod(method) {
	case "express":
		if subtotal >= 80000 {
			return 1200
		}
		return 2500
	default:
		return shippingFeeForSubtotal(subtotal)
	}
}

func discountForCoupon(code string, subtotal int64) int64 {
	switch normalizeCouponCode(code, "") {
	case "SAVE10":
		return subtotal / 10
	case "WELCOME500":
		if subtotal >= 5000 {
			return 500
		}
	case "FREESHIP":
		return 0
	}
	if subtotal > 20000 {
		return 1000
	}
	return 0
}

func normalizeCouponCode(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.ToUpper(strings.TrimSpace(primary))
	}
	return strings.ToUpper(strings.TrimSpace(fallback))
}

func primaryImage(images []domain.ProductImage) string {
	for _, image := range images {
		if image.Primary {
			return image.URL
		}
	}
	if len(images) > 0 {
		return images[0].URL
	}
	return ""
}

func coalesceAddress(candidate, fallback domain.Address) domain.Address {
	if candidate.Line1 == "" && candidate.City == "" && candidate.CountryCode == "" {
		return fallback
	}
	return candidate
}

func newOrderNumber() string {
	return "ORD-" + time.Now().UTC().Format("20060102-150405")
}

func (s *CheckoutService) calculateCartSnapshot(ctx context.Context, items []domain.CartItem) (int64, string, error) {
	var subtotal int64
	currency := "USD"

	for _, item := range items {
		product, found, err := s.products.FindByID(ctx, item.ProductID)
		if err != nil {
			return 0, "", err
		}
		if !found || !product.Active || product.Stock < item.Quantity {
			return 0, "", ErrValidation
		}
		subtotal += product.Price * int64(item.Quantity)
		currency = product.Currency
	}

	return subtotal, currency, nil
}
