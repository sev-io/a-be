package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/middleware"
	"vilow-be/pkg/models"
	"vilow-be/pkg/utils"
	"vilow-be/prisma/db"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error generating password hash: %v", err), http.StatusInternalServerError)
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.Email.Equals(user.Email),
		).Exec(r.Context())

		if err == nil && existingUser != nil {
			http.Error(w, "E-mail already in use", http.StatusConflict)
			return
		}

		createdUser, err := client.User.CreateOne(
			db.User.Name.Set(user.Name),
			db.User.Email.Set(user.Email),
			db.User.Password.Set(string(hashedPassword)),
			db.User.StrID.Set(user.StrID),
			db.User.Description.Set(user.Description),
		).Exec(r.Context())

		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating a new user: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "User created! ID: %s", createdUser.ID)
	}
}

func UpdateUserHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			log.Println("AuthContext not found in context")
			return
		}

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
			log.Printf("Error decoding request body: %v\n", err)
			return
		}

		if !utils.ValidateUser(&user) {
			http.Error(w, "Invalid user data", http.StatusBadRequest)
			log.Println("Invalid user data")
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.ID.Equals(authContext.UserID),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
			log.Println("User not found: " + err.Error())
			return
		}

		updateData, err := utils.BuildUpdateData(&user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error building update data: %v", err), http.StatusInternalServerError)
			log.Printf("Error building update data: %v\n", err)
			return
		} else if len(updateData) == 0 {
			http.Error(w, "No data to update", http.StatusBadRequest)
			log.Println("No data to update")
			return
		}

		_, err = client.User.FindUnique(
			db.User.ID.Equals(existingUser.ID),
		).Update(
			updateData...,
		).Exec(r.Context())

		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
			log.Printf("Error updating user: %v\n", err)
			return
		}

		fmt.Fprintf(w, "User updated! ID: %s", existingUser.ID)
	}
}

func DeleteUserHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			log.Println("AuthContext not found in context")
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.ID.Equals(authContext.UserID),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
			log.Println("User not found: " + err.Error())
			return
		}

		_, err = client.User.FindUnique(
			db.User.ID.Equals(existingUser.ID),
		).Delete().Exec(r.Context())

		if err != nil {
			http.Error(w, fmt.Sprintf("Error deleting user: %v", err), http.StatusInternalServerError)
			log.Printf("Error deleting user: %v\n", err)
			return
		}

		fmt.Fprintf(w, "User deleted! ID: %s", existingUser.ID)
	}
}

func GetUserDataHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			log.Println("AuthContext not found in context")
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		existingUser, err := client.User.FindUnique(
			db.User.StrID.Equals(id),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
			log.Println("User not found: " + err.Error())
			return
		}

		response, err := utils.BuildResponse(existingUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error building response: %v", err), http.StatusInternalServerError)
			log.Printf("Error building response: %v\n", err)
			return
		}

		utils.SendResponse(w, response)
	}
}
