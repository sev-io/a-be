package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"vilow-be/pkg/models"
	"vilow-be/prisma/db"

	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
)

type AuthContext struct {
	UserID string
}

// func HomeHandler(client *db.PrismaClient) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		authContext, ok := r.Context().Value("authContext").(AuthContext)
// 		if !ok {
// 			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
// 			return
// 		}

// 		existingUser, err := client.User.FindUnique(
// 			db.User.ID.Equals(authContext.UserID),
// 		).Exec(r.Context())

// 		if err != nil || existingUser == nil {
// 			http.Error(w, "User not found", http.StatusUnauthorized)
// 			return
// 		}

// 		// Adicione a lógica de paginação aqui
// 		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
// 		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

// 		if page < 1 {
// 			page = 1
// 		}

// 		if limit < 1 || limit > 50 {
// 			limit = 10
// 		}

// 		skip := (page - 1) * limit

// 		medias, err := client.Media.FindMany({
// 			skip: skip,
// 			take: limit,
// 		}, nil
// 		).Exec(r.Context())

// 		if err != nil {
// 			http.Error(w, "Error fetching medias", http.StatusInternalServerError)
// 			return
// 		}

// 		// Converta as medias para JSON e retorne como resposta
// 		jsonData, err := json.Marshal(medias)
// 		if err != nil {
// 			http.Error(w, "Error converting medias to JSON", http.StatusInternalServerError)
// 			return
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write(jsonData)
// 	}
// }

func ProductHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := r.Context().Value("authContext").(AuthContext)
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

		videoList, err := client.Media.FindMany().Exec(r.Context())

		if err != nil {
			http.Error(w, "Error fetching data", http.StatusUnauthorized)
			return
		}

		var response = struct {
			UserName  string
			UserEmail string
			Videos    []db.MediaModel
		}{
			UserName:  existingUser.Name,
			UserEmail: existingUser.Email,
			Videos:    videoList,
		}

		// Convertendo a struct para JSON
		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Erro ao converter para JSON:", err)
			return
		}

		// Convertendo bytes para string para exibição
		jsonString := string(jsonData)

		fmt.Fprintf(w, "%s", jsonString)
	}
}

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
			http.Error(w, "Error generating password hash", http.StatusInternalServerError)
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
			http.Error(w, "Error creating a new user", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "User created! ID: %s", createdUser.ID)
	}
}

func UploadHandler(client *db.PrismaClient, minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		bucketName := os.Getenv("BUCKET_NAME")

		authContext, ok := r.Context().Value("authContext").(AuthContext)
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

		err = r.ParseMultipartForm(900 << 20)
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
		json.NewEncoder(w).Encode(createdMedia)
	}
}
