package middleware

import (
	"context"
	"net/http"
	"strings"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/utils"
	"vilow-be/prisma/db"
)

type AuthContextKey string

func AuthMiddleware(next http.HandlerFunc, client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		userId := r.Header.Get("UserId")

		if authHeader == "" {
			http.Error(w, "Token not given", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, "Bearer ")
		if len(tokenString) != 2 {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		token, err := utils.VerifyToken(tokenString[1], userId)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		user, err := client.User.FindUnique(
			db.User.ID.Equals(userId),
		).Exec(r.Context())

		if err != nil || user == nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusUnauthorized)
			return
		}

		authContext := dto.AuthContext{
			UserID:   userId,
			Name:     user.Name,
			Email:    user.Email,
			StrID:    user.StrID,
			Subjects: user.Subjects,
		}

		ctx := context.WithValue(r.Context(), AuthContextKey("authContext"), authContext)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Replace '*' with specific origin
		// TODO: Replace 'POST, GET, OPTIONS, PUT, DELETE' with the methods the API supports
		// TODO: Replace 'Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization' with the headers the API expects
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, r)
	})
}
