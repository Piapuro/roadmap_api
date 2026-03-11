package requests

type SignUpRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8"  example:"password123"`
	Name     string `json:"name"     validate:"required"        example:"山田太郎"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required"        example:"password123"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required" example:"山田花子"`
}
