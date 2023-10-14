package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/handlers"
	"vilow-be/pkg/middleware"
	"vilow-be/pkg/utils"
	"vilow-be/prisma/db"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetUserDataHandler(t *testing.T) {
	r := mux.NewRouter()

	dir, _ := os.Getwd()
	if err := godotenv.Load(filepath.Join(dir, "../../.env")); err != nil {
		t.Fatalf("Failed loading .env file: %v", err)
	}

	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		t.Fatalf("Failed to create Prisma client: %v", err)
	}

	defer func() {
		_ = client.Prisma.Disconnect()
	}()

	r.HandleFunc("/user/{id}", handlers.GetUserDataHandler(client))

	req, err := http.NewRequest("GET", "/user/test-id", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthContextKey("authContext"), dto.AuthContext{UserID: "test-id"})
	req = req.WithContext(ctx)

	token, _ := utils.GenerateToken("test-id")
	req.Header.Add("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be OK")
	assert.True(t, strings.Contains(rr.Body.String(), "User updated! ID: test-id"), "Response body should contain user ID")
}
