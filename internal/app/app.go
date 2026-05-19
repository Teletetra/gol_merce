package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"ecommerce_go/internal/config"
	"ecommerce_go/internal/domain"
	"ecommerce_go/internal/httpapi"
	"ecommerce_go/internal/httpapi/middleware"
	"ecommerce_go/internal/service"
	"ecommerce_go/internal/store/memory"
)

func Run() error {
	cfg := config.Load()
	store := memory.NewStore()

	authService := service.NewAuthService(store, cfg.TokenSecret, cfg.TokenTTL)
	productService := service.NewProductService(memoryProductRepo{store})
	cartService := service.NewCartService(store, memoryProductRepo{store})
	checkoutService := service.NewCheckoutService(
		store,
		memoryProductRepo{store},
		memoryOrderRepo{store},
		memoryPaymentRepo{store},
		service.MockGateway{Provider: cfg.PaymentProvider},
	)

	if _, _, err := authService.Register(context.Background(), service.RegisterInput{
		Name:     "Platform Admin",
		Email:    cfg.AdminEmail,
		Password: cfg.AdminPassword,
	}, domain.RoleAdmin); err != nil && err != service.ErrConflict {
		return err
	}

	if err := seedProducts(context.Background(), productService); err != nil {
		return err
	}

	handlers := httpapi.NewHandlers(authService, productService, cartService, checkoutService)
	router := routes(handlers, authService)

	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("e-commerce backend listening on %s", cfg.Address())
	return server.ListenAndServe()
}

func routes(h *httpapi.Handlers, auth *service.AuthService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /v1/auth/register", h.Register)
	mux.HandleFunc("POST /v1/auth/login", h.Login)
	mux.HandleFunc("GET /v1/products", h.ListProducts)
	mux.HandleFunc("GET /v1/products/{id}", h.GetProduct)

	protected := middleware.RequireAuth(auth)
	admin := chain(protected, middleware.RequireRole(domain.RoleAdmin))

	mux.Handle("GET /v1/auth/me", protected(http.HandlerFunc(h.Me)))
	mux.Handle("GET /v1/cart", protected(http.HandlerFunc(h.GetCart)))
	mux.Handle("PUT /v1/cart/items", protected(http.HandlerFunc(h.UpsertCartItem)))
	mux.Handle("DELETE /v1/cart", protected(http.HandlerFunc(h.ClearCart)))
	mux.Handle("POST /v1/checkout", protected(http.HandlerFunc(h.Checkout)))
	mux.Handle("GET /v1/orders", protected(http.HandlerFunc(h.ListOrders)))

	mux.Handle("POST /v1/admin/products", admin(http.HandlerFunc(h.CreateProduct)))
	mux.Handle("PUT /v1/admin/products/{id}", admin(http.HandlerFunc(h.UpdateProduct)))
	mux.Handle("DELETE /v1/admin/products/{id}", admin(http.HandlerFunc(h.DeleteProduct)))

	return middleware.Recover(middleware.CORS(middleware.Logger(mux)))
}

func chain(mw ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			next = mw[i](next)
		}
		return next
	}
}

func seedProducts(ctx context.Context, products *service.ProductService) error {
	seeds := []service.ProductInput{
		{
			Name:        "AeroFit Pro Runner",
			Description: "Responsive performance sneakers with breathable knit mesh.",
			Price:       12999,
			Currency:    "USD",
			Stock:       60,
			Category:    "footwear",
			Tags:        []string{"sports", "running", "new"},
			Active:      true,
		},
		{
			Name:        "Nimbus Smart Watch",
			Description: "AMOLED smartwatch with health tracking and NFC payments.",
			Price:       24999,
			Currency:    "USD",
			Stock:       35,
			Category:    "wearables",
			Tags:        []string{"smart", "fitness"},
			Active:      true,
		},
		{
			Name:        "Studio Desk Lamp",
			Description: "Minimal task lamp with wireless charging base.",
			Price:       8999,
			Currency:    "USD",
			Stock:       80,
			Category:    "home-office",
			Tags:        []string{"decor", "workspace"},
			Active:      true,
		},
	}

	for _, seed := range seeds {
		if _, err := products.Create(ctx, seed); err != nil {
			return err
		}
	}
	return nil
}

type memoryProductRepo struct{ *memory.Store }
type memoryOrderRepo struct{ *memory.Store }
type memoryPaymentRepo struct{ *memory.Store }

func (r memoryProductRepo) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	return r.Store.CreateProduct(ctx, product)
}

func (r memoryProductRepo) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	return r.Store.UpdateProduct(ctx, product)
}

func (r memoryProductRepo) Delete(ctx context.Context, id string) error {
	return r.Store.Delete(ctx, id)
}

func (r memoryProductRepo) FindByID(ctx context.Context, id string) (domain.Product, bool, error) {
	return r.Store.FindProductByID(ctx, id)
}

func (r memoryProductRepo) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	return r.Store.List(ctx, filter)
}

func (r memoryOrderRepo) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	return r.Store.CreateOrder(ctx, order)
}

func (r memoryOrderRepo) FindByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	return r.Store.FindByUser(ctx, userID)
}

func (r memoryOrderRepo) FindByID(ctx context.Context, id string) (domain.Order, bool, error) {
	return r.Store.FindOrderByID(ctx, id)
}

func (r memoryOrderRepo) Update(ctx context.Context, order domain.Order) (domain.Order, error) {
	return r.Store.UpdateOrder(ctx, order)
}

func (r memoryPaymentRepo) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	return r.Store.CreatePayment(ctx, payment)
}

func (r memoryPaymentRepo) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	return r.Store.UpdatePayment(ctx, payment)
}

func (r memoryPaymentRepo) FindByID(ctx context.Context, id string) (domain.Payment, bool, error) {
	return r.Store.FindPaymentByID(ctx, id)
}
