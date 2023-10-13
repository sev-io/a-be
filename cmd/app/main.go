package main

import (
	"log"
	"net/http"
	"vilow-be/config"

	"github.com/joho/godotenv"
)

var PORT = ":8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	client, err := config.SetupDatabase()
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
	defer client.Prisma.Disconnect()

	minioClient, err := config.SetupMinio()
	if err != nil {
		log.Fatalf("Error setting up MinIO: %v", err)
	}

	corsHandler := config.SetupServer(client, minioClient)

	log.Printf("Server running on port %s", PORT)
	log.Fatal(http.ListenAndServe(PORT, corsHandler))
}
