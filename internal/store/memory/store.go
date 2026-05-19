package memory

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"ecommerce_go/internal/domain"
)

type Store struct {
	mu       sync.RWMutex
	users    map[string]domain.User
	products map[string]domain.Product
	carts    map[string]domain.Cart
	orders   map[string]domain.Order
	payments map[string]domain.Payment
}

func NewStore() *Store {
	return &Store{
		users:    make(map[string]domain.User),
		products: make(map[string]domain.Product),
		carts:    make(map[string]domain.Cart),
		orders:   make(map[string]domain.Order),
		payments: make(map[string]domain.Payment),
	}
}

func (s *Store) Create(ctx context.Context, user domain.User) (domain.User, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
	return user, nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.Email == email {
			return user, true, nil
		}
	}
	return domain.User{}, false, nil
}

func (s *Store) FindByID(ctx context.Context, id string) (domain.User, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok, nil
}

func (s *Store) CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.products[product.ID] = product
	return product, nil
}

func (s *Store) UpdateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[product.ID]; !ok {
		return domain.Product{}, errors.New("product not found")
	}
	s.products[product.ID] = product
	return product, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.products, id)
	return nil
}

func (s *Store) FindProductByID(ctx context.Context, id string) (domain.Product, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, ok := s.products[id]
	return product, ok, nil
}

func (s *Store) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]domain.Product, 0, len(s.products))
	for _, product := range s.products {
		if filter.Active != nil && product.Active != *filter.Active {
			continue
		}
		if filter.Category != "" && !strings.EqualFold(product.Category, filter.Category) {
			continue
		}
		if filter.Query != "" {
			haystack := strings.ToLower(product.Name + " " + product.Description + " " + product.Category)
			if !strings.Contains(haystack, strings.ToLower(filter.Query)) {
				continue
			}
		}
		products = append(products, product)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].CreatedAt.After(products[j].CreatedAt)
	})

	start := filter.Offset
	if start > len(products) {
		start = len(products)
	}

	end := start + filter.Limit
	if filter.Limit <= 0 || end > len(products) {
		end = len(products)
	}

	return products[start:end], nil
}

func (s *Store) GetByUserID(ctx context.Context, userID string) (domain.Cart, error) {
	_ = ctx
	s.mu.RLock()
	cart, ok := s.carts[userID]
	s.mu.RUnlock()
	if ok {
		return cart, nil
	}
	return domain.Cart{UserID: userID, Items: []domain.CartItem{}, UpdatedAt: time.Now().UTC()}, nil
}

func (s *Store) Save(ctx context.Context, cart domain.Cart) (domain.Cart, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.carts[cart.UserID] = cart
	return cart, nil
}

func (s *Store) Clear(ctx context.Context, userID string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.carts, userID)
	return nil
}

func (s *Store) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
	return order, nil
}

func (s *Store) FindByUser(ctx context.Context, userID string) ([]domain.Order, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	orders := make([]domain.Order, 0)
	for _, order := range s.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

func (s *Store) FindOrderByID(ctx context.Context, id string) (domain.Order, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	return order, ok, nil
}

func (s *Store) UpdateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
	return order, nil
}

func (s *Store) CreatePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[payment.ID] = payment
	return payment, nil
}

func (s *Store) UpdatePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[payment.ID] = payment
	return payment, nil
}

func (s *Store) FindPaymentByID(ctx context.Context, id string) (domain.Payment, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	payment, ok := s.payments[id]
	return payment, ok, nil
}
