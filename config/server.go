package config

import (
	"net/http"
	"vilow-be/pkg/handlers"
	"vilow-be/pkg/middleware"
	"vilow-be/prisma/db"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/rs/cors"
)

// SetupServer is a function that sets up the server
func SetupServer(client *db.PrismaClient, minioClient *minio.Client) http.Handler {
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowCredentials: true,
	})

	// r.Use(middleware.CorsMiddleware)
	r.HandleFunc("/user", handlers.CreateUserHandler(client)).Methods(http.MethodPost)
	r.HandleFunc("/login", handlers.AuthHandler(client)).Methods(http.MethodPost)

	protectedRouter := r.PathPrefix("/in").Subrouter()
	protectedRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(middleware.AuthMiddleware(next.ServeHTTP))
	})
	protectedRouter.HandleFunc("/upload", handlers.UploadHandler(client, minioClient)).Methods(http.MethodPost)

	return c.Handler(r)
}
