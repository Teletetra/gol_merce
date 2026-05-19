package domain

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Currency    string    `json:"currency"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	UnitPrice   int64  `json:"unit_price"`
	Quantity    int    `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
}

type Address struct {
	FullName    string `json:"full_name"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	Phone       string `json:"phone"`
}

type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Items           []OrderItem `json:"items"`
	Subtotal        int64       `json:"subtotal"`
	Tax             int64       `json:"tax"`
	ShippingFee     int64       `json:"shipping_fee"`
	Discount        int64       `json:"discount"`
	GrandTotal      int64       `json:"grand_total"`
	Currency        string      `json:"currency"`
	Status          OrderStatus `json:"status"`
	ShippingAddress Address     `json:"shipping_address"`
	PaymentID       string      `json:"payment_id"`
	CreatedAt       time.Time   `json:"created_at"`
}

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusCaptured PaymentStatus = "captured"
	PaymentStatusFailed   PaymentStatus = "failed"
)

type Payment struct {
	ID         string        `json:"id"`
	OrderID    string        `json:"order_id"`
	Provider   string        `json:"provider"`
	Method     string        `json:"method"`
	Amount     int64         `json:"amount"`
	Currency   string        `json:"currency"`
	Status     PaymentStatus `json:"status"`
	Reference  string        `json:"reference"`
	FailureMsg string        `json:"failure_message,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}
