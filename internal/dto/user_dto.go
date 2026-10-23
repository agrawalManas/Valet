package dto

type UpdateUserRequest struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
	Role     *string `json:"role"`
}

func (r *UpdateUserRequest) IsEmpty() bool {
	return r.FullName == nil && r.Email == nil && r.Role == nil
}
