package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ValidateTokenResponse struct {
	UserID string `json:"user_id"`
	Valid  bool   `json:"valid"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
