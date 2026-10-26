package dto

type UpdateProfileRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,min=1"`
	Email    *string `json:"email"     binding:"omitempty"`
	Gender   *string `json:"gender" binding:"omitempty,oneof=MALE FEMALE OTHER"`
	Dob      *string `json:"dob" binding:"omitempty,datetime=2006-01-02"`
}

func (r *UpdateProfileRequest) IsEmpty() bool {
	return r.FullName == nil && r.Email == nil && r.Gender == nil && r.Dob == nil
}

type AdminUpdateUserRequest struct {
	Role     *string `json:"role"      binding:"omitempty,oneof=CUSTOMER DRIVER SUPERVISOR"`
	IsActive *bool   `json:"is_active"`
	TargetId string  `json:"target_id" binding:"required"`
}

func (r *AdminUpdateUserRequest) IsValid() bool { return r.Role == nil && r.IsActive == nil }
