package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce_go/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("resource not found")
	ErrValidation         = errors.New("validation failed")
	ErrConflict           = errors.New("conflict")
	ErrPaymentFailed      = errors.New("payment failed")
)

type tokenClaims struct {
	UserID string      `json:"uid"`
	Role   domain.Role `json:"role"`
	Exp    int64       `json:"exp"`
}

func newID(prefix string) string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(buf))
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func verifyPassword(rawPassword, hashed string) bool {
	return hashPassword(rawPassword) == hashed
}

func signToken(secret string, claims tokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + signature, nil
}

func parseToken(secret, token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return tokenClaims{}, ErrUnauthorized
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return tokenClaims{}, ErrUnauthorized
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenClaims{}, ErrUnauthorized
	}

	var claims tokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return tokenClaims{}, ErrUnauthorized
	}

	if claims.Exp < time.Now().Unix() {
		return tokenClaims{}, ErrUnauthorized
	}

	return claims, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, value)
	return strings.Trim(value, "-")
}
