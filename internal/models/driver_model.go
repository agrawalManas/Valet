package models

import (
	"time"
)

type Driver struct {
	Id                string         `json:"id"`
	DriverDetails     *User          `json:"driver_details"`
	SiteId            string         `json:"site_id"`
	AddedBy           string         `json:"added_by"`
	CurrentSupervisor string         `json:"current_supervisor"`
	ShiftStartTime    time.Time      `json:"shift_start_time"`
	ShiftEndTime      time.Time      `json:"shift_end_time"`
	CurrentStatus     string         `json:"current_status"`
	LastLogin         *time.Time     `json:"last_login"`
	LastLogout        *time.Time     `json:"last_logout"`
	TotalTipEarned    float64        `json:"total_tip_earned"`
	KYCStatus         string         `json:"kyc_status"`
	AadhaarNumber     string         `json:"aadhaar_number"`
	AadhaarUrl        string         `json:"aadhaar_url"`
	PanNumber         string         `json:"pan_number"`
	PanUrl            string         `json:"pan_url"`
	DrivingLicense    string         `json:"driving_license"`
	DrivingLicenseUrl string         `json:"driving_license_url"`
	TotalExperience   int16          `json:"total_experience"`
	Reviews           *driverReviews `json:"reviews"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type driverReviews struct {
	AvgRating float32   `json:"avg_rating"`
	Reviews   []reviews `json:"reviews"`
}

type reviews struct {
	Id         string    `json:"id"`
	DriverId   string    `json:"-"`
	ReviewerId string    `json:"reviewer_id"`
	Rating     float32   `json:"rating"`
	Review     string    `json:"review"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
