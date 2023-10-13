package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/models"
	"vilow-be/prisma/db"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func AuthHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.Email.Equals(user.Email),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
		if err != nil {
			http.Error(w, "Wrong password", http.StatusUnauthorized)
			return
		}

		tokenString, err := GenerateToken(existingUser.ID)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: time.Now().Add(5 * time.Hour),
		})

		response := &dto.LoginResponse{
			AuthToken: tokenString,
			UserID:    existingUser.ID,
			UserEmail: existingUser.Email,
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Error encoding the response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(responseBytes))
	}
}

func VerifyToken(tokenString string, userID string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expTime := time.Unix(int64(claims["exp"].(float64)), 0)
		if time.Now().UTC().After(expTime) {
			newTokenString, err := GenerateToken(userID)
			if err != nil {
				return nil, err
			}

			newToken, err := jwt.Parse(newTokenString, func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtKey, nil
			})

			if err != nil {
				return nil, err
			}

			return newToken, nil
		}
	}

	return token, nil
}

func GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Hour)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
