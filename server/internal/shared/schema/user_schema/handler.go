package userschema

type SignupRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email,max=255"`
	// bcrypt silently truncates/errors past 72 bytes, so cap it here.
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type SignupResponse struct {
	Message string `json:"message"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	Message     string `json:"message"`
}

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
	Message     string `json:"message"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type LogoutAllResponse struct {
	Message string `json:"message"`
}
