package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"ecommerce_go/internal/domain"
	"ecommerce_go/internal/service"
)

type Handlers struct {
	auth     *service.AuthService
	products *service.ProductService
	cart     *service.CartService
	checkout *service.CheckoutService
}

func NewHandlers(
	auth *service.AuthService,
	products *service.ProductService,
	cart *service.CartService,
	checkout *service.CheckoutService,
) *Handlers {
	return &Handlers{auth: auth, products: products, cart: cart, checkout: checkout}
}

func (h *Handlers) Health(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, token, err := h.auth.Register(r.Context(), input, domain.RoleCustomer)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{"user": user, "token": token})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, token, err := h.auth.Login(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"user": user, "token": token})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "missing user context")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter := domain.ProductFilter{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		Limit:    intQuery(r, "limit", 20),
		Offset:   intQuery(r, "offset", 0),
	}
	if rawActive := r.URL.Query().Get("active"); rawActive != "" {
		active := rawActive == "true"
		filter.Active = &active
	}

	products, err := h.products.List(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"items": products, "count": len(products)})
}

func (h *Handlers) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	product, err := h.products.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, product)
}

func (h *Handlers) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input service.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	product, err := h.products.Create(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, product)
}

func (h *Handlers) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var input service.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	id := r.PathValue("id")
	product, err := h.products.Update(r.Context(), id, input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, product)
}

func (h *Handlers) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.products.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetCart(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	cart, err := h.cart.Get(r.Context(), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, cart)
}

func (h *Handlers) UpsertCartItem(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	var payload struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	cart, err := h.cart.UpsertItem(r.Context(), user.ID, payload.ProductID, payload.Quantity)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, cart)
}

func (h *Handlers) ClearCart(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	if err := h.cart.Clear(r.Context(), user.ID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Checkout(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	var input service.CheckoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	order, payment, err := h.checkout.Checkout(r.Context(), user.ID, input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{"order": order, "payment": payment})
}

func (h *Handlers) ListOrders(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	orders, err := h.checkout.ListOrders(r.Context(), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": orders})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials), errors.Is(err, service.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrPaymentFailed):
		WriteError(w, http.StatusPaymentRequired, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func intQuery(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
