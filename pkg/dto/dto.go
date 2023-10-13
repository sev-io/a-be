package dto

type LoginResponse struct {
	AuthToken string `json:"authToken"`
	UserID    string `json:"userId"`
	UserEmail string `json:"userEmail"`
}
