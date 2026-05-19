package service

import (
	"context"
	"time"

	"ecommerce_go/internal/domain"
)

type CartService struct {
	carts    domain.CartRepository
	products domain.ProductRepository
}

func NewCartService(carts domain.CartRepository, products domain.ProductRepository) *CartService {
	return &CartService{carts: carts, products: products}
}

func (s *CartService) Get(ctx context.Context, userID string) (domain.Cart, error) {
	cart, err := s.carts.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	return s.enrichCart(ctx, cart)
}

func (s *CartService) UpsertItem(ctx context.Context, userID, productID string, quantity int) (domain.Cart, error) {
	if quantity < 0 {
		return domain.Cart{}, ErrValidation
	}

	product, found, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return domain.Cart{}, err
	}
	if !found || !product.Active {
		return domain.Cart{}, ErrNotFound
	}
	if quantity > product.Stock {
		return domain.Cart{}, ErrValidation
	}

	cart, err := s.carts.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}

	nextItems := make([]domain.CartItem, 0, len(cart.Items)+1)
	replaced := false
	for _, item := range cart.Items {
		if item.ProductID == productID {
			replaced = true
			if quantity > 0 {
				nextItems = append(nextItems, domain.CartItem{ProductID: productID, Quantity: quantity})
			}
			continue
		}
		nextItems = append(nextItems, item)
	}

	if !replaced && quantity > 0 {
		nextItems = append(nextItems, domain.CartItem{ProductID: productID, Quantity: quantity})
	}

	cart.Items = nextItems
	cart.UpdatedAt = time.Now().UTC()
	cart, err = s.enrichCart(ctx, cart)
	if err != nil {
		return domain.Cart{}, err
	}
	return s.carts.Save(ctx, cart)
}

func (s *CartService) Clear(ctx context.Context, userID string) error {
	return s.carts.Clear(ctx, userID)
}

func (s *CartService) ApplyCoupon(ctx context.Context, userID, couponCode string) (domain.Cart, error) {
	cart, err := s.carts.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	cart.CouponCode = normalizeCouponCode(couponCode, "")
	cart.UpdatedAt = time.Now().UTC()
	cart, err = s.enrichCart(ctx, cart)
	if err != nil {
		return domain.Cart{}, err
	}
	return s.carts.Save(ctx, cart)
}

func (s *CartService) enrichCart(ctx context.Context, cart domain.Cart) (domain.Cart, error) {
	var subtotal int64
	currency := "USD"
	items := make([]domain.CartItem, 0, len(cart.Items))

	for _, item := range cart.Items {
		product, found, err := s.products.FindByID(ctx, item.ProductID)
		if err != nil {
			return domain.Cart{}, err
		}
		if !found {
			continue
		}

		enriched := item
		enriched.ProductName = product.Name
		enriched.ProductSlug = product.Slug
		enriched.SKU = product.SKU
		enriched.UnitPrice = product.Price
		enriched.Subtotal = product.Price * int64(item.Quantity)
		if len(product.Images) > 0 {
			enriched.ImageURL = product.Images[0].URL
		}

		subtotal += enriched.Subtotal
		currency = product.Currency
		items = append(items, enriched)
	}

	discount := discountForCoupon(cart.CouponCode, subtotal)
	tax := subtotal / 10
	shipping := shippingFeeForSubtotal(subtotal)
	cart.Items = items
	cart.Subtotal = subtotal
	cart.Tax = tax
	cart.ShippingFee = shipping
	cart.Discount = discount
	cart.GrandTotal = subtotal + tax + shipping - discount
	cart.Currency = currency
	return cart, nil
}
