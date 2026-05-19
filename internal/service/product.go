package service

import (
	"context"
	"strings"
	"time"

	"ecommerce_go/internal/domain"
)

type ProductService struct {
	products domain.ProductRepository
}

func NewProductService(products domain.ProductRepository) *ProductService {
	return &ProductService{products: products}
}

type ProductInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int64    `json:"price"`
	Currency    string   `json:"currency"`
	Stock       int      `json:"stock"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Active      bool     `json:"active"`
}

func (s *ProductService) Create(ctx context.Context, input ProductInput) (domain.Product, error) {
	if strings.TrimSpace(input.Name) == "" || input.Price <= 0 || input.Stock < 0 {
		return domain.Product{}, ErrValidation
	}

	now := time.Now().UTC()
	product := domain.Product{
		ID:          newID("prd"),
		Name:        strings.TrimSpace(input.Name),
		Slug:        slugify(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
		Currency:    defaultCurrency(input.Currency),
		Stock:       input.Stock,
		Category:    strings.TrimSpace(input.Category),
		Tags:        input.Tags,
		Active:      input.Active,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.products.Create(ctx, product)
}

func (s *ProductService) Update(ctx context.Context, id string, input ProductInput) (domain.Product, error) {
	product, found, err := s.products.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	if !found {
		return domain.Product{}, ErrNotFound
	}

	if strings.TrimSpace(input.Name) == "" || input.Price <= 0 || input.Stock < 0 {
		return domain.Product{}, ErrValidation
	}

	product.Name = strings.TrimSpace(input.Name)
	product.Slug = slugify(input.Name)
	product.Description = strings.TrimSpace(input.Description)
	product.Price = input.Price
	product.Currency = defaultCurrency(input.Currency)
	product.Stock = input.Stock
	product.Category = strings.TrimSpace(input.Category)
	product.Tags = input.Tags
	product.Active = input.Active
	product.UpdatedAt = time.Now().UTC()

	return s.products.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.products.Delete(ctx, id)
}

func (s *ProductService) Get(ctx context.Context, id string) (domain.Product, error) {
	product, found, err := s.products.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	if !found {
		return domain.Product{}, ErrNotFound
	}
	return product, nil
}

func (s *ProductService) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	return s.products.List(ctx, filter)
}

func defaultCurrency(currency string) string {
	if strings.TrimSpace(currency) == "" {
		return "USD"
	}
	return strings.ToUpper(strings.TrimSpace(currency))
}
