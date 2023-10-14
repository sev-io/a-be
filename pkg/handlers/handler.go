package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/middleware"
	"vilow-be/prisma/db"
)

func FeedHandler(client *db.PrismaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := r.Context().Value(middleware.AuthContextKey("authContext")).(dto.AuthContext)
		if !ok {
			http.Error(w, "AuthContext not found in context", http.StatusInternalServerError)
			return
		}

		videoList, err := client.Media.FindMany().Exec(r.Context())

		if err != nil {
			http.Error(w, "Error fetching data", http.StatusUnauthorized)
			return
		}

		response := &dto.FeedResponse{
			UserAuthData: authContext,
			Medias:       videoList,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Erro ao converter para JSON:", err)
			return
		}

		jsonString := string(jsonData)

		fmt.Fprintf(w, "%s", jsonString)
	}
}
