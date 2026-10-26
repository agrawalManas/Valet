package models

import (
	"time"
)

type User struct {
	Id           string    `json:"id"`
	FullName     *string   `json:"full_name"`
	Email        *string   `json:"email"`
	MobileNumber string    `json:"mobile_number"`
	Role         *string   `json:"-"`
	IsActive     bool      `json:"is_active"`
	Gender       *string   `json:"gender"`
	Dob          *string   `json:"dob"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
