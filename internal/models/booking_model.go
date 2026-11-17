package models

import "time"

type Booking struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	CreatedBy string `json:"created_by"`
	SiteId    string `json:"site_id"`
	SlotId    string `json:"slot_id"`

	VehicleNumber string `json:"vehicle_number"`
	VehicleName   string `json:"vehicle_name"`

	BookingDate time.Time `json:"booking_date"`
	Status      string    `json:"status"`
	BookingType string    `json:"booking_type"`
	BookingMode string    `json:"booking_mode"`

	ValetAmount     float64 `json:"valet_amount"`
	InsuranceOpted  bool    `json:"insurance_opted"`
	InsuranceAmount float64 `json:"insurance_amount"`
	TotalTax        float64 `json:"total_tax"`
	TotalAmount     float64 `json:"total_amount"`

	PaymentMethod *string `json:"payment_method"`
	PaymentStatus *string `json:"payment_status"`

	CancelledAt  *time.Time `json:"cancelled_at"`
	CancelReason *string    `json:"cancel_reason"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
