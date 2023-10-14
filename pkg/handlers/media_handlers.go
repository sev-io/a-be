package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"vilow-be/pkg/dto"
	"vilow-be/prisma/db"

	"github.com/minio/minio-go/v7"
)

func UploadMediaHandler(client *db.PrismaClient, minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		bucketName := os.Getenv("BUCKET_NAME")

		authContext, ok := r.Context().Value("authContext").(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.ID.Equals(authContext.UserID),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		err = r.ParseMultipartForm(1000 << 20)
		if err != nil {
			http.Error(w, "Não foi possível processar o corpo da requisição", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Não foi possível obter o arquivo do formulário", http.StatusBadRequest)
			return
		}
		defer file.Close()

		name := r.FormValue("name")
		description := r.FormValue("description")

		var contentAfterLastDot string
		lastDotIndex := strings.LastIndex(handler.Filename, ".")
		if lastDotIndex != -1 && lastDotIndex < len(handler.Filename)-1 {
			contentAfterLastDot = handler.Filename[lastDotIndex+1:]
		}
		now := time.Now()
		formattedTime := now.Format("02012006-150405")
		objectName := handler.Filename + "video_" + formattedTime + "." + contentAfterLastDot

		contentType := handler.Header.Get("Content-Type")
		savedFile, err := minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			fmt.Printf("Erro ao fazer upload do vídeo para o MinIO: %s", err.Error())
			http.Error(w, "Erro ao fazer upload do vídeo para o MinIO", http.StatusInternalServerError)
			return
		}

		createdMedia, err := client.Media.CreateOne(
			db.Media.Name.Set(name),
			db.Media.Path.Set(savedFile.Location),
			db.Media.Description.Set(description),
			db.Media.User.Link(
				db.User.ID.Equals(existingUser.ID),
			),
		).Exec(r.Context())

		if err != nil {
			http.Error(w, "Erro ao criar o vídeo no banco de dados", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(createdMedia)

		if err != nil {
			http.Error(w, "Erro ao converter o vídeo para JSON", http.StatusInternalServerError)
			return
		}
	}
}
