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

	// Public routes
	// User public routes
	r.HandleFunc("/user", handlers.CreateUserHandler(client)).Methods(http.MethodPost)

	// Auth public routes
	r.HandleFunc("/login", handlers.AuthHandler(client)).Methods(http.MethodPost)

	// Protected routes
	protectedRouter := r.PathPrefix("/in").Subrouter()
	protectedRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(middleware.AuthMiddleware(next.ServeHTTP, client))
	})

	// User protected routes
	protectedRouter.HandleFunc("/user", handlers.UpdateUserHandler(client)).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/user", handlers.DeleteUserHandler(client)).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/{id}", handlers.GetUserDataHandler(client)).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/", handlers.FeedHandler(client)).Methods(http.MethodGet)

	// Media protected routes
	protectedRouter.HandleFunc("/media/upload", handlers.UploadMediaHandler(client, minioClient)).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/media/{id}", handlers.GetMediaHandler(client)).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/media/{id}", handlers.UpdateMediaHandler(client, minioClient)).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/media/{id}", handlers.DeleteMediaHandler(client, minioClient)).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/medias/timeline", handlers.GetMediasTimelineHandler(client)).Methods(http.MethodGet)

	return c.Handler(r)
}
