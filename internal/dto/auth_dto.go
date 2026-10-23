package dto

type VerifyOTPRequest struct {
	Otp          string  `json:"otp" binding:"required"`
	MobileNumber string  `json:"mobile_number" binding:"required"`
	Role         *string `json:"role" binding:"omitempty"`
}
