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
	return s.carts.GetByUserID(ctx, userID)
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
	return s.carts.Save(ctx, cart)
}

func (s *CartService) Clear(ctx context.Context, userID string) error {
	return s.carts.Clear(ctx, userID)
}
