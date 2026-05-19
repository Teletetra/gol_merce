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

type ProductImage struct {
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	Primary   bool   `json:"primary"`
	SortOrder int    `json:"sort_order"`
}

type ProductSEO struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type ProductRating struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type Product struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Slug           string            `json:"slug"`
	SKU            string            `json:"sku"`
	Brand          string            `json:"brand"`
	Description    string            `json:"description"`
	ShortDesc      string            `json:"short_description"`
	Price          int64             `json:"price"`
	CompareAtPrice int64             `json:"compare_at_price"`
	Currency       string            `json:"currency"`
	Stock          int               `json:"stock"`
	Category       string            `json:"category"`
	Tags           []string          `json:"tags"`
	Images         []ProductImage    `json:"images"`
	Attributes     map[string]string `json:"attributes"`
	SEO            ProductSEO        `json:"seo"`
	Rating         ProductRating     `json:"rating"`
	Featured       bool              `json:"featured"`
	Active         bool              `json:"active"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type CartItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name,omitempty"`
	ProductSlug string `json:"product_slug,omitempty"`
	SKU         string `json:"sku,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	UnitPrice   int64  `json:"unit_price,omitempty"`
	Quantity    int    `json:"quantity"`
	Subtotal    int64  `json:"subtotal,omitempty"`
}

type Cart struct {
	UserID      string     `json:"user_id"`
	Items       []CartItem `json:"items"`
	CouponCode  string     `json:"coupon_code,omitempty"`
	Subtotal    int64      `json:"subtotal"`
	Tax         int64      `json:"tax"`
	ShippingFee int64      `json:"shipping_fee"`
	Discount    int64      `json:"discount"`
	GrandTotal  int64      `json:"grand_total"`
	Currency    string     `json:"currency"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type OrderItem struct {
	ProductID   string            `json:"product_id"`
	ProductName string            `json:"product_name"`
	ProductSlug string            `json:"product_slug"`
	SKU         string            `json:"sku"`
	ImageURL    string            `json:"image_url"`
	UnitPrice   int64             `json:"unit_price"`
	Quantity    int               `json:"quantity"`
	Discount    int64             `json:"discount"`
	Subtotal    int64             `json:"subtotal"`
	Attributes  map[string]string `json:"attributes,omitempty"`
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
	ID              string       `json:"id"`
	OrderNumber     string       `json:"order_number"`
	UserID          string       `json:"user_id"`
	Items           []OrderItem  `json:"items"`
	Subtotal        int64        `json:"subtotal"`
	Tax             int64        `json:"tax"`
	ShippingFee     int64        `json:"shipping_fee"`
	Discount        int64        `json:"discount"`
	GrandTotal      int64        `json:"grand_total"`
	Currency        string       `json:"currency"`
	Status          OrderStatus  `json:"status"`
	CouponCode      string       `json:"coupon_code,omitempty"`
	ShippingMethod  string       `json:"shipping_method"`
	Notes           string       `json:"notes,omitempty"`
	ShippingAddress Address      `json:"shipping_address"`
	BillingAddress  Address      `json:"billing_address"`
	PaymentID       string       `json:"payment_id"`
	Events          []OrderEvent `json:"events"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	PaidAt          *time.Time   `json:"paid_at,omitempty"`
}

type OrderEvent struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

type Payment struct {
	ID            string            `json:"id"`
	OrderID       string            `json:"order_id"`
	Provider      string            `json:"provider"`
	Method        string            `json:"method"`
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	Status        PaymentStatus     `json:"status"`
	Reference     string            `json:"reference"`
	ProviderTxnID string            `json:"provider_txn_id"`
	FailureMsg    string            `json:"failure_message,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	ProcessedAt   *time.Time        `json:"processed_at,omitempty"`
}
