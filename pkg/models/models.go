package models

import "time"

type User struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	StrID         string         `json:"strId"`
	Description   string         `json:"description"`
	Medias        []Media        `json:"medias"`
	Password      string         `json:"password"`
	Followers     []Follow       `json:"followers"`
	Following     []Follow       `json:"following"`
	Notifications []Notification `json:"notifications"`
	Likes         []Like         `json:"likes"`
	Dislikes      []Dislike      `json:"dislikes"`
	Comments      []Comment      `json:"comments"`
	Subjects      []string       `json:"subjects"`
}

type Media struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Subjects    []string  `json:"subjects"`
	User        User      `json:"user"`
	UserID      string    `json:"userId"`
	Likes       []Like    `json:"likes"`
	Dislikes    []Dislike `json:"dislikes"`
	Comments    []Comment `json:"comments"`
}

type Follow struct {
	ID        string `json:"id"`
	Follower  User   `json:"follower"`
	Following User   `json:"following"`
}

type Notification struct {
	ID        string    `json:"id"`
	User      User      `json:"user"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
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
