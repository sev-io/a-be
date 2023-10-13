package middleware

import (
	"context"
	"net/http"
	"strings"
	"vilow-be/pkg/handlers"
)

type AuthContextKey string

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		userId := r.Header.Get("UserId")
		if authHeader == "" {
			http.Error(w, "Token not given", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, "Bearer ")
		if len(tokenString) != 2 {
			http.Error(w, "Invalid Token", http.StatusBadRequest)
			return
		}

		token, err := handlers.VerifyToken(tokenString[1], userId)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		authContext := handlers.AuthContext{
			UserID: userId,
		}

		ctx := context.WithValue(r.Context(), AuthContextKey("authContext"), authContext)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, r)
	})
}