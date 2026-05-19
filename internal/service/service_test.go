package service

import (
	"context"
	"testing"
	"time"

	"ecommerce_go/internal/domain"
	"ecommerce_go/internal/store/memory"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	store := memory.NewStore()
	auth := NewAuthService(store, "secret", time.Hour)

	user, token, err := auth.Register(context.Background(), RegisterInput{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: "Password123",
	}, domain.RoleCustomer)
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if user.Email != "jane@example.com" {
		t.Fatalf("unexpected email: %s", user.Email)
	}
	if token == "" {
		t.Fatal("expected token to be returned")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token returned error: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("unexpected user id in token: got %s want %s", claims.UserID, user.ID)
	}

	_, loginToken, err := auth.Login(context.Background(), LoginInput{
		Email:    "jane@example.com",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if loginToken == "" {
		t.Fatal("expected login token")
	}
}

func TestCheckoutRestoresStockOnPaymentFailure(t *testing.T) {
	store := memory.NewStore()
	productRepo := testProductRepo{store}
	orderRepo := testOrderRepo{store}
	paymentRepo := testPaymentRepo{store}
	cartService := NewCartService(store, productRepo)
	checkoutService := NewCheckoutService(store, productRepo, orderRepo, paymentRepo, MockGateway{Provider: "mockpay"})

	product, err := productRepo.Create(context.Background(), domain.Product{
		ID:        "prd_1",
		Name:      "Demo Product",
		Slug:      "demo-product",
		Price:     10000,
		Currency:  "USD",
		Stock:     5,
		Category:  "demo",
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create product returned error: %v", err)
	}

	if _, err := cartService.UpsertItem(context.Background(), "usr_1", product.ID, 2); err != nil {
		t.Fatalf("upsert cart item returned error: %v", err)
	}

	order, payment, err := checkoutService.Checkout(context.Background(), "usr_1", CheckoutInput{
		PaymentMethod: "card",
		CardToken:     "fail_card",
		CustomerEmail: "jane@example.com",
		ShippingAddress: domain.Address{
			FullName:    "Jane Doe",
			Line1:       "1 Main St",
			City:        "Austin",
			State:       "TX",
			PostalCode:  "73301",
			CountryCode: "US",
			Phone:       "+1-555-0100",
		},
	})
	if err == nil {
		t.Fatal("expected checkout to fail")
	}
	if order.Status != domain.OrderStatusCancelled {
		t.Fatalf("unexpected order status: %s", order.Status)
	}
	if payment.Status != domain.PaymentStatusFailed {
		t.Fatalf("unexpected payment status: %s", payment.Status)
	}

	updatedProduct, found, err := productRepo.FindByID(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("find product returned error: %v", err)
	}
	if !found {
		t.Fatal("expected product to still exist")
	}
	if updatedProduct.Stock != 5 {
		t.Fatalf("expected stock rollback to 5, got %d", updatedProduct.Stock)
	}
}

type testProductRepo struct{ *memory.Store }
type testOrderRepo struct{ *memory.Store }
type testPaymentRepo struct{ *memory.Store }

func (r testProductRepo) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	return r.Store.CreateProduct(ctx, product)
}

func (r testProductRepo) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	return r.Store.UpdateProduct(ctx, product)
}

func (r testProductRepo) Delete(ctx context.Context, id string) error {
	return r.Store.Delete(ctx, id)
}

func (r testProductRepo) FindByID(ctx context.Context, id string) (domain.Product, bool, error) {
	return r.Store.FindProductByID(ctx, id)
}

func (r testProductRepo) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	return r.Store.List(ctx, filter)
}

func (r testOrderRepo) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	return r.Store.CreateOrder(ctx, order)
}

func (r testOrderRepo) FindByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	return r.Store.FindByUser(ctx, userID)
}

func (r testOrderRepo) FindByID(ctx context.Context, id string) (domain.Order, bool, error) {
	return r.Store.FindOrderByID(ctx, id)
}

func (r testOrderRepo) Update(ctx context.Context, order domain.Order) (domain.Order, error) {
	return r.Store.UpdateOrder(ctx, order)
}

func (r testPaymentRepo) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	return r.Store.CreatePayment(ctx, payment)
}

func (r testPaymentRepo) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	return r.Store.UpdatePayment(ctx, payment)
}

func (r testPaymentRepo) FindByID(ctx context.Context, id string) (domain.Payment, bool, error) {
	return r.Store.FindPaymentByID(ctx, id)
}
