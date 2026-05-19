package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"ecommerce_go/internal/domain"
	"ecommerce_go/internal/httpapi"
	"ecommerce_go/internal/service"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, string(debug.Stack()))
				httpapi.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			token := strings.TrimPrefix(raw, "Bearer ")
			if token == "" || token == raw {
				httpapi.WriteError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := auth.ValidateToken(token)
			if err != nil {
				httpapi.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			user, err := auth.Me(r.Context(), claims.UserID)
			if err != nil {
				httpapi.WriteError(w, http.StatusUnauthorized, "user no longer exists")
				return
			}

			next.ServeHTTP(w, r.WithContext(httpapi.WithUser(r.Context(), user)))
		})
	}
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := httpapi.UserFromContext(r.Context())
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "missing user context")
				return
			}
			if user.Role != role {
				httpapi.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
