package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	TokenSecret     string
	TokenTTL        time.Duration
	AdminEmail      string
	AdminPassword   string
	PaymentProvider string
}

func Load() Config {
	return Config{
		Port:            getEnv("APP_PORT", "8080"),
		TokenSecret:     getEnv("TOKEN_SECRET", "change-me-super-secret"),
		TokenTTL:        getDurationEnv("TOKEN_TTL_HOURS", 24) * time.Hour,
		AdminEmail:      getEnv("ADMIN_EMAIL", "admin@shop.local"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", "Admin123!"),
		PaymentProvider: getEnv("PAYMENT_PROVIDER", "mockpay"),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback int) time.Duration {
	raw := getEnv(key, "")
	if raw == "" {
		return time.Duration(fallback)
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return time.Duration(fallback)
	}
	return time.Duration(value)
}
