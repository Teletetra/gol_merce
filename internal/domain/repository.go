package domain

import "context"

type UserRepository interface {
	Create(context.Context, User) (User, error)
	FindByEmail(context.Context, string) (User, bool, error)
	FindByID(context.Context, string) (User, bool, error)
}

type ProductRepository interface {
	Create(context.Context, Product) (Product, error)
	Update(context.Context, Product) (Product, error)
	Delete(context.Context, string) error
	FindByID(context.Context, string) (Product, bool, error)
	List(context.Context, ProductFilter) ([]Product, error)
}

type CartRepository interface {
	GetByUserID(context.Context, string) (Cart, error)
	Save(context.Context, Cart) (Cart, error)
	Clear(context.Context, string) error
}

type OrderRepository interface {
	Create(context.Context, Order) (Order, error)
	FindByUserID(context.Context, string) ([]Order, error)
	FindByID(context.Context, string) (Order, bool, error)
	Update(context.Context, Order) (Order, error)
}

type PaymentRepository interface {
	Create(context.Context, Payment) (Payment, error)
	Update(context.Context, Payment) (Payment, error)
	FindByID(context.Context, string) (Payment, bool, error)
}

type ProductFilter struct {
	Query    string
	Category string
	Active   *bool
	Limit    int
	Offset   int
}
