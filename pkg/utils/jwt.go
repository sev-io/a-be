package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func VerifyToken(tokenString string, userID string) (*jwt.Token, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		return nil, errors.New("JWT_SECRET_KEY is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func GenerateToken(userID string) (string, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		return "", errors.New("JWT_SECRET_KEY is not set")
	}

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
