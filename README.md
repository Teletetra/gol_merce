# Modern E-Commerce Backend in Go

This project provides a production-style e-commerce backend written in Go with:

- JWT-like HMAC token authentication
- Product catalog management with admin-only CRUD
- Customer cart management
- Checkout and order creation
- Mock payment gateway integration
- Layered services, repositories, and HTTP middleware

## Run

```powershell
go run ./cmd/api
```

The API starts on `http://localhost:8080`.

## Default Admin

- Email: `admin@shop.local`
- Password: `Admin123!`

Override with environment variables:

```powershell
$env:APP_PORT="8080"
$env:TOKEN_SECRET="replace-me"
$env:TOKEN_TTL_HOURS="24"
$env:ADMIN_EMAIL="admin@shop.local"
$env:ADMIN_PASSWORD="Admin123!"
$env:PAYMENT_PROVIDER="mockpay"
go run ./cmd/api
```

## Main Endpoints

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `GET /v1/auth/me`
- `GET /v1/products`
- `GET /v1/products/{id}`
- `POST /v1/admin/products`
- `PUT /v1/admin/products/{id}`
- `DELETE /v1/admin/products/{id}`
- `GET /v1/cart`
- `PUT /v1/cart/items`
- `DELETE /v1/cart`
- `POST /v1/checkout`
- `GET /v1/orders`

## Example Flow

Register:

```json
POST /v1/auth/register
{
  "name": "Ava Stone",
  "email": "ava@example.com",
  "password": "Password123"
}
```

Add item to cart:

```json
PUT /v1/cart/items
{
  "product_id": "prd_xxx",
  "quantity": 2
}
```

Checkout:

```json
POST /v1/checkout
{
  "payment_method": "card",
  "card_token": "tok_demo_visa",
  "customer_email": "ava@example.com",
  "shipping_address": {
    "full_name": "Ava Stone",
    "line1": "221B Market Street",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country_code": "US",
    "phone": "+1-555-0100"
  }
}
```
