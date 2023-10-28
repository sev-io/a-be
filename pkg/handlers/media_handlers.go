package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/middleware"
	"vilow-be/prisma/db"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

func UploadMediaHandler(client *db.PrismaClient, minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		bucketName := os.Getenv("BUCKET_NAME")

		authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
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
			http.Error(w, "Unable to process request body", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Unable to get file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		name := r.FormValue("name")
		description := r.FormValue("description")
		subjects := r.Form["subjects"]

		re := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
		sanitizedFilename := re.ReplaceAllString(handler.Filename, "_")

		var contentAfterLastDot string
		lastDotIndex := strings.LastIndex(sanitizedFilename, ".")
		if lastDotIndex != -1 && lastDotIndex < len(sanitizedFilename)-1 {
			contentAfterLastDot = sanitizedFilename[lastDotIndex+1:]
		}
		now := time.Now()
		formattedTime := now.Format("02012006-150405")
		objectName := sanitizedFilename + "video_" + formattedTime + "." + contentAfterLastDot

		contentType := handler.Header.Get("Content-Type")
		savedFile, err := minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			fmt.Printf("Error uploading video to MinIO: %s", err.Error())
			http.Error(w, "Error uploading video to MinIO", http.StatusInternalServerError)
			return
		}

		createdMedia, err := client.Media.CreateOne(
			db.Media.Name.Set(name),
			db.Media.Path.Set(savedFile.Location),
			db.Media.Description.Set(description),
			db.Media.User.Link(
				db.User.ID.Equals(existingUser.ID),
			),
			db.Media.Subjects.Set(subjects),
		).Exec(r.Context())

		if err != nil {
			http.Error(w, "Error creating video in the database", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(createdMedia)

		if err != nil {
			http.Error(w, "Error converting video to JSON", http.StatusInternalServerError)
			return
		}
	}
}

func GetMediaHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		mediaID := vars["id"]

		media, err := client.Media.FindUnique(
			db.Media.ID.Equals(mediaID),
		).With(
			db.Media.Likes.Fetch(),
			db.Media.Dislikes.Fetch(),
			db.Media.Comments.Fetch(),
		).Exec(r.Context())

		if err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(media)
		if err != nil {
			http.Error(w, "Error converting media to JSON", http.StatusInternalServerError)
			return
		}
	}
}

func UpdateMediaHandler(client *db.PrismaClient, minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()
		bucketName := os.Getenv("BUCKET_NAME")

		vars := mux.Vars(r)
		mediaID := vars["id"]

		_, media, errStatusCode, err := getMediaAndAuthContext(r, client, mediaID)
		if err != nil {
			http.Error(w, err.Error(), errStatusCode)
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")
		subjects := r.Form["subjects"]

		file, handler, err := r.FormFile("video")
		if err == nil {
			defer file.Close()

			re := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
			sanitizedFilename := re.ReplaceAllString(handler.Filename, "_")

			var contentAfterLastDot string
			lastDotIndex := strings.LastIndex(sanitizedFilename, ".")
			if lastDotIndex != -1 && lastDotIndex < len(sanitizedFilename)-1 {
				contentAfterLastDot = sanitizedFilename[lastDotIndex+1:]
			}
			now := time.Now()
			formattedTime := now.Format("02012006-150405")
			objectName := sanitizedFilename + "video_" + formattedTime + "." + contentAfterLastDot

			contentType := handler.Header.Get("Content-Type")
			_, err = minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				fmt.Printf("Error uploading video to MinIO: %s", err.Error())
				http.Error(w, "Error uploading video to MinIO", http.StatusInternalServerError)
				return
			}

			media.Path = objectName
		}

		updatedMedia, err := client.Media.FindUnique(
			db.Media.ID.Equals(mediaID),
		).Update(
			db.Media.Name.Set(name),
			db.Media.Description.Set(description),
			db.Media.Subjects.Set(subjects),
			db.Media.Path.Set(media.Path),
		).Exec(r.Context())

		if err != nil {
			http.Error(w, "Error updating media in the database", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(updatedMedia)
		if err != nil {
			http.Error(w, "Error converting media to JSON", http.StatusInternalServerError)
			return
		}
	}
}

func DeleteMediaHandler(client *db.PrismaClient, minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaID := mux.Vars(r)["id"]
		bucketName := os.Getenv("BUCKET_NAME")

		_, media, errStatusCode, err := getMediaAndAuthContext(r, client, mediaID)

		if err != nil {
			http.Error(w, err.Error(), errStatusCode)
			return
		}

		objectName := strings.TrimPrefix(media.Path, "http://localhost:9000/vilow-videos/")
		err = minioClient.RemoveObject(r.Context(), bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			http.Error(w, "Error deleting media file from MinIO: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = client.Like.FindMany(
			db.Like.MediaID.Equals(mediaID),
		).Delete().Exec(r.Context())
		if err != nil {
			http.Error(w, "Error deleting likes associated with media", http.StatusInternalServerError)
			return
		}

		_, err = client.Dislike.FindMany(
			db.Dislike.MediaID.Equals(mediaID),
		).Delete().Exec(r.Context())
		if err != nil {
			http.Error(w, "Error deleting dislikes associated with media", http.StatusInternalServerError)
			return
		}

		_, err = client.Comment.FindMany(
			db.Comment.MediaID.Equals(mediaID),
		).Delete().Exec(r.Context())
		if err != nil {
			http.Error(w, "Error deleting comments associated with media", http.StatusInternalServerError)
			return
		}

		_, err = client.Media.FindUnique(
			db.Media.ID.Equals(mediaID),
		).Delete().Exec(r.Context())
		if err != nil {
			http.Error(w, "Error deleting media from database", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Media and associated data successfully deleted"))
	}
}

func GetMediasTimelineHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			return
		}

		existingUser, err := client.User.FindUnique(
			db.User.ID.Equals(authContext.UserID),
		).Exec(r.Context())

		if err != nil || existingUser == nil {
			http.Error(w, "User not found: "+existingUser.Email, http.StatusUnauthorized)
			return
		}

		pageSize := 10
		lastMediaID := r.URL.Query().Get("lastMediaID")

		var medias []db.MediaModel

		if lastMediaID != "" {
			medias, err = client.
				Media.
				FindMany(
					db.Media.Subjects.HasSome(existingUser.Subjects),
					db.Media.ID.Gt(lastMediaID),
				).
				OrderBy(
					db.Media.ID.Order(db.ASC),
				).
				Take(pageSize).
				Cursor(
					db.Media.ID.Cursor(lastMediaID),
				).
				Exec(ctx)
		} else {
			medias, err = client.
				Media.
				FindMany(
					db.Media.Subjects.HasSome(existingUser.Subjects),
				).
				OrderBy(
					db.Media.ID.Order(db.ASC),
				).
				Take(pageSize).
				Exec(ctx)
		}

		if err != nil {
			http.Error(w, "Error fetching medias", http.StatusInternalServerError)
			return
		}

		if len(medias) == 0 && lastMediaID != "" {
			lastMediaID = ""
			medias, err = client.
				Media.
				FindMany().
				Take(pageSize).
				Skip(0).
				OrderBy(
					db.Media.ID.Order(db.ASC),
				).Exec(ctx)

			if err != nil {
				http.Error(w, "Error fetching medias", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(medias)

		if err != nil {
			http.Error(w, "Error converting medias to JSON", http.StatusInternalServerError)
			return
		}
	}
}
