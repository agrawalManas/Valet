package models

import (
	"time"
)

type Site struct {
	Id               string    `json:"id"`
	SiteName         string    `json:"site_name"`
	IsBookingAllowed bool      `json:"is_booking_allowed"`
	ParkingType      string    `json:"parking_type"`
	TotalCapacity    int       `json:"total_capacity"`
	DriverCount      int       `json:"driver_count"`
	OpeningTime      time.Time `json:"site_opening_time"`
	ClosingTime      time.Time `json:"site_closing_time"`
	ValetFee         float64   `json:"valet_fee"`
	SiteImages       []string  `json:"site_images"`

	SiteAddress       SiteAddress `json:"site_address"`
	SupervisorId      string      `json:"supervisor_id"`
	SupervisorDetails User        `json:"supervisor_details"`

	InsuranceAddOnAllowed bool                `json:"insurance_add_on_allowed"`
	SiteInsuranceAddOn    *SiteInsuranceAddOn `json:"site_insurance_add_on,omitempty"`
	TippingAllowed        bool                `json:"tipping_allowed"`
	TippingDetails        *TippingDetails     `json:"tipping_details,omitempty"`

	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteAddress struct {
	PinCode         string  `json:"pin_code"`
	City            string  `json:"city"`
	State           string  `json:"state"`
	AddressLine1    string  `json:"address_line_1"`
	AddressLine2    *string `json:"address_line_2,omitempty"`
	SiteCoordinates *string `json:"site_coordinates,omitempty"`
}

type SiteInsuranceAddOn struct {
	InsuranceAddOnPrice *float64 `json:"insurance_add_on_price,omitempty"`
}

type TippingDetails struct {
	TippingAmount      []float64 `json:"tipping_amount"`
	PickTippingAllowed *bool     `json:"pick_tipping_allowed,omitempty"`
	DropTippingAllowed *bool     `json:"drop_tipping_allowed,omitempty"`
}
