package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"vilow-be/pkg/dto"
	"vilow-be/pkg/models"
	"vilow-be/prisma/db"

	"golang.org/x/crypto/bcrypt"
)

func ValidateUser(user *models.User) bool {
	if user.Email != "" {
		_, err := mail.ParseAddress(user.Email)
		if err != nil {
			log.Printf("Error parsing email address: %v\n", err)
			return false
		}
	}
	return true
}

func BuildUpdateData(user *models.User) (updateData []db.UserSetParam, err error) {
	if user.Name != "" {
		updateData = append(updateData, db.User.Name.Set(user.Name))
	}

	if user.Email != "" {
		updateData = append(updateData, db.User.Email.Set(user.Email))
	}

	if user.Password != "" {
		var hashedPassword []byte
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error generating password hash: %v\n", err)
			return nil, err
		}

		updateData = append(updateData, db.User.Password.Set(string(hashedPassword)))
	}

	if user.StrID != "" {
		updateData = append(updateData, db.User.StrID.Set(user.StrID))
	}

	if user.Description != "" {
		updateData = append(updateData, db.User.Description.Set(user.Description))
	}

	return updateData, nil
}

func BuildResponse(existingUser *db.UserModel) (*dto.User, error) {
	response := &dto.User{
		ID:          existingUser.ID,
		Name:        existingUser.Name,
		Email:       existingUser.Email,
		StrID:       existingUser.StrID,
		Description: existingUser.Description,
		Medias:      make([]dto.Media, len(existingUser.Medias())),
	}

	for i, media := range existingUser.Medias() {
		response.Medias[i] = dto.Media{
			ID:          media.ID,
			Name:        media.Name,
			Path:        media.Path,
			Description: media.Description,
			UserID:      media.UserID,
			Likes:       make([]dto.Like, len(media.Likes())),
			Comments:    make([]dto.Comment, len(media.Comments())),
		}

		for j, like := range media.Likes() {
			response.Medias[i].Likes[j] = dto.Like{
				ID:    like.ID,
				User:  dto.User{ID: like.User().ID},
				Media: dto.Media{ID: like.Media().ID},
			}
		}

		for k, comment := range media.Comments() {
			response.Medias[i].Comments[k] = dto.Comment{
				ID:      comment.ID,
				User:    dto.User{ID: comment.User().ID},
				Media:   dto.Media{ID: comment.Media().ID},
				Content: comment.Content,
			}
		}
	}

	return response, nil
}

func SendResponse(w http.ResponseWriter, response *dto.User) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("Error converting to JSON: %v\n", err)
		return
	}

	jsonString := string(jsonData)

	fmt.Fprintf(w, "%s", jsonString)
}
