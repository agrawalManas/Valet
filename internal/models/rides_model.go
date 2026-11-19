package models

import "time"

type Ride struct {
	Id              string   `json:"id"`
	BookingId       string   `json:"booking_id"`
	RideStatus      string   `json:"ride_status"`
	CreatedBy       string   `json:"created_by"`
	PickDriverId    string   `json:"-"`
	DropDriverId    string   `json:"-"`
	SupervisorId    string   `json:"-"`
	PickReviewId    string   `json:"pick_review_id"`
	DropReviewId    string   `json:"drop_review_id"`
	RideReviewId    string   `json:"ride_review_id"`
	ParkingLocation string   `json:"parking_location"`
	ParkedCarImages []string `json:"parked_car_image"`
	Otp             int16    `json:"otp"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type RideReview struct {
	Id         string    `json:"id"`
	ReviewerId string    `json:"reviewer_id"`
	Rating     float32   `json:"rating"`
	Review     string    `json:"review"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RideHistory struct {
	RideId         string    `json:"ride_id"`
	PreviousStatus *string   `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	ActorId        *string   `json:"-"`
	ActorRole      *string   `json:"-"`
	DriverId       *string   `json:"driver_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
