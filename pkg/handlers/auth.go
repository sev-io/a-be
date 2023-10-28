package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/middleware"
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
			http.Error(w, "User not found: "+err.Error(), http.StatusUnauthorized)
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

func getMediaAndAuthContext(r *http.Request, client *db.PrismaClient, mediaID string) (dto.AuthContext, *db.MediaModel, int, error) {
	authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
	if !ok {
		return dto.AuthContext{}, nil, http.StatusUnauthorized, errors.New("AuthContext not found in context")
	}

	media, err := client.Media.FindUnique(
		db.Media.ID.Equals(mediaID),
	).Exec(r.Context())

	if err != nil {
		return dto.AuthContext{}, nil, http.StatusUnauthorized, errors.New("media not found")
	}

	if media.UserID != authContext.UserID {
		return dto.AuthContext{}, nil, http.StatusUnauthorized, errors.New("unauthorized: You do not have permission to manipulate this media")
	}

	return authContext, media, http.StatusOK, nil
}
