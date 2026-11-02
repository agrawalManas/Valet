package dto

type FileUploadDTO struct {
	Visibility *string `form:"visibility" json:"visibility" binding:"omitempty,oneof=PUBLIC PRIVATE"`
}
