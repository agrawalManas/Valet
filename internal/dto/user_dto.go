package dto

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required"`
}
