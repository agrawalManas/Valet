package dto

import (
	"fmt"
	"time"
)

type CreateDriverRequest struct {
	FullName     string  `json:"full_name" binding:"required"`
	MobileNumber string  `json:"mobile_number" binding:"required"`
	Email        *string `json:"email"     binding:"omitempty"`
	Gender       string  `json:"gender" binding:"required,oneof=MALE FEMALE OTHER"`
	Dob          string  `json:"dob" binding:"required,datetime=2006-01-02"`

	SiteId string `json:"site_id" binding:"required"`

	ShiftStartTime time.Time `json:"shift_start_time" binding:"required"`
	ShiftEndTime   time.Time `json:"shift_end_time" binding:"required"`

	AadhaarNumber     string `json:"aadhaar_number" binding:"required"`
	AadhaarUrl        string `json:"aadhaar_url" binding:"required"`
	PanNumber         string `json:"pan_number" binding:"required"`
	PanUrl            string `json:"pan_url" binding:"required"`
	DrivingLicense    string `json:"driving_license" binding:"required"`
	DrivingLicenseUrl string `json:"driving_license_url" binding:"required"`

	TotalExperience int16 `json:"total_experience" binding:"required"`
}

type UpdateDriverRequest struct {
	Id              string     `json:"id" binding:"required"`
	Status          *string    `json:"status" binding:"omitempty,oneof=AVAILABLE ON_DUTY OFF_DUTY SUSPENDED INACTIVE"`
	ShiftStartTime  *time.Time `json:"shift_start_time" binding:"omitempty"`
	ShiftEndTime    *time.Time `json:"shift_end_time" binding:"omitempty"`
	TotalExperience *int16     `json:"total_experience" binding:"omitempty"`
}

type ChangeSupervisorForDriverRequest struct {
	DriverId        string `json:"driver_id" binding:"required"`
	NewSuperVisorId string `json:"new_supervisor_id" binding:"required"`
}

func (req *UpdateDriverRequest) ValidateUpdateDriverRequest() error {
	if req.Status == nil && req.ShiftStartTime == nil && req.ShiftEndTime == nil && req.TotalExperience == nil {
		return fmt.Errorf("Invalid request body: Provide one of status, shift_start_time, shift_end_time, total_experience")
	}
	return nil
}
