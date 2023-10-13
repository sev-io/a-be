package config

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// SetupMinio is a function that sets up the MinIO client
func SetupMinio() (*minio.Client, error) {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT_URL")
	minioAccessKey := os.Getenv("MINIO_ROOT_USER")
	minioSecretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	useSSL := false

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}
