package dto

import "vilow-be/prisma/db"

type LoginResponse struct {
	AuthToken string `json:"authToken"`
	UserID    string `json:"userId"`
	UserEmail string `json:"userEmail"`
}

type AuthContext struct {
	UserID   string   `json:"userId"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	StrID    string   `json:"strId"`
	Subjects []string `json:"subjects"`
}

type FeedResponse struct {
	UserAuthData AuthContext     `json:"userAuthData"`
	Medias       []db.MediaModel `json:"medias"`
}

type Media struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Subjects    []string  `json:"subjects"`
	UserID      string    `json:"userId"`
	Likes       []Like    `json:"likes"`
	Dislikes    []Dislike `json:"dislikes"`
	Comments    []Comment `json:"comments"`
}

type Like struct {
	ID    string `json:"id"`
	User  User   `json:"user"`
	Media Media  `json:"media"`
}

type Dislike struct {
	ID    string `json:"id"`
	User  User   `json:"user"`
	Media Media  `json:"media"`
}

type Comment struct {
	ID      string `json:"id"`
	User    User   `json:"user"`
	Media   Media  `json:"media"`
	Content string `json:"content"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	StrID string `json:"strId"`
}
