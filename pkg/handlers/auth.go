package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/models"
	"vilow-be/pkg/utils"
	"vilow-be/prisma/db"

	"golang.org/x/crypto/bcrypt"
)

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

		tokenString, err := utils.GenerateToken(existingUser.ID)
		if err != nil {
			http.Error(w, fmt.Errorf(`error generating token: %v`, err).Error(), http.StatusInternalServerError)
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
