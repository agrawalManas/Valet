package dto

import (
	"errors"
	"time"
)

type AddSiteRequest struct {
	SiteName         string    `json:"site_name"          binding:"required,max=255"`
	IsBookingAllowed *bool     `json:"is_booking_allowed" binding:"required"`
	ParkingType      string    `json:"parking_type"       binding:"required,oneof=BASEMENT GROUND OTHER"`
	TotalCapacity    int       `json:"total_capacity"     binding:"required,min=1"`
	DriverCount      int       `json:"driver_count"       binding:"min=0"`
	OpeningTime      time.Time `json:"site_opening_time"  binding:"required"`
	ClosingTime      time.Time `json:"site_closing_time"  binding:"required"`
	ValetFee         float64   `json:"valet_fee"          binding:"min=0"`
	SiteImages       []string  `json:"site_images"        binding:"omitempty,max=10,dive,url"`

	SiteAddress       SiteAddressInput `json:"site_address"`
	SupervisorDetails SupervisorInput  `json:"supervisor_details"`

	InsuranceAddOnAllowed *bool           `json:"insurance_add_on_allowed" binding:"required"`
	SiteInsuranceAddOn    *InsuranceInput `json:"site_insurance_add_on"`
	TippingAllowed        *bool           `json:"tipping_allowed"          binding:"required"`
	TippingDetails        *TippingInput   `json:"tipping_details"`
}

type SiteAddressInput struct {
	PinCode         string  `json:"pin_code"       binding:"required,len=6,numeric"`
	City            string  `json:"city"           binding:"required,max=100"`
	State           string  `json:"state"          binding:"required,max=100"`
	AddressLine1    string  `json:"address_line_1" binding:"required,max=255"`
	AddressLine2    *string `json:"address_line_2" binding:"omitempty,max=255"`
	SiteCoordinates *string `json:"site_coordinates" binding:"omitempty"`
}

type SupervisorInput struct {
	FullName     string  `json:"full_name" binding:"required,max=255"`
	Email        *string `json:"email"`
	MobileNumber string  `json:"mobile_number" binding:"required,len=10,numeric"`
	Dob          string  `json:"dob"           binding:"required"`
	Gender       string  `json:"gender"        binding:"required,oneof=MALE FEMALE OTHER"`
}

type InsuranceInput struct {
	Price *float64 `json:"insurance_add_on_price" binding:"omitempty"`
}

type TippingInput struct {
	Amounts            []float64 `json:"tipping_amount" binding:"required,min=1,max=10"`
	PickTippingAllowed *bool     `json:"pick_tipping_allowed"`
	DropTippingAllowed *bool     `json:"drop_tipping_allowed"`
}

func (r *AddSiteRequest) Validate() error {
	if !r.OpeningTime.Before(r.ClosingTime) {
		return errors.New("site_closing_time must be after site_opening_time")
	}
	if *r.InsuranceAddOnAllowed && (r.SiteInsuranceAddOn == nil || r.SiteInsuranceAddOn.Price == nil) {
		return errors.New("insurance price is required when insurance add-on is allowed")
	}
	if *r.TippingAllowed && r.TippingDetails == nil {
		return errors.New("tipping_details is required when tipping is allowed")
	}

	return nil
}

type GetSiteRequest struct {
	CreatedBy    *string `form:"created_by"    json:"created_by"    binding:"omitempty"`
	SupervisorId *string `form:"supervisor_id" json:"supervisor_id" binding:"omitempty"`
	ShowAll      *bool   `form:"show_all"      json:"show_all"`
}

func (r *GetSiteRequest) Validate() error {
	selected := 0
	if r.CreatedBy != nil {
		selected++
	}
	if r.SupervisorId != nil {
		selected++
	}
	if r.ShowAll != nil && *r.ShowAll { // show_all=false does not count as a filter
		selected++
	}

	switch {
	case selected == 0:
		return errors.New("provide exactly one of created_by, supervisor_id or show_all=true")
	case selected > 1:
		return errors.New("only one of created_by, supervisor_id or show_all can be provided at a time")
	}
	return nil
}

// Every field is optional. A field that is not sent stays nil, and the SQL
// keeps the stored value with COALESCE($n, column). Because "not sent" and
// "send null" both end up as nil, a nullable column cannot be cleared through
// this request (address_line_2 and the insurance price are examples).
//
// Validation tags use `omitempty` first so they only run when a value is sent.
// For pointer fields a sent zero value still counts as present, so
// "site_name": "" fails min=1 and "total_capacity": 0 fails min=1.
type UpdateSiteRequest struct {
	SiteName         *string    `json:"site_name"          binding:"omitempty,min=1,max=255"`
	IsBookingAllowed *bool      `json:"is_booking_allowed"`
	ParkingType      *string    `json:"parking_type"       binding:"omitempty,oneof=BASEMENT GROUND OTHER"`
	TotalCapacity    *int       `json:"total_capacity"     binding:"omitempty,min=1"`
	DriverCount      *int       `json:"driver_count"       binding:"omitempty,min=0"`
	OpeningTime      *time.Time `json:"site_opening_time"`
	ClosingTime      *time.Time `json:"site_closing_time"`
	ValetFee         *float64   `json:"valet_fee"          binding:"omitempty,min=0"`

	// nil = keep the stored images, [] = replace them with an empty list.
	SiteImages []string `json:"site_images" binding:"omitempty,max=10,dive,url"`

	// Only the address fields that are sent are changed.
	SiteAddress *UpdateSiteAddressInput `json:"site_address"`

	// If sent, this must be complete (all required fields of SupervisorInput).
	// The service then finds or creates that supervisor, as CreateSite does.
	SupervisorMobileNumber string `json:"supervisor_mobile_number" binding:"required,len=10"`

	InsuranceAddOnAllowed *bool                 `json:"insurance_add_on_allowed"`
	SiteInsuranceAddOn    *UpdateInsuranceInput `json:"site_insurance_add_on"`
	TippingAllowed        *bool                 `json:"tipping_allowed"`
	TippingDetails        *UpdateTippingInput   `json:"tipping_details"`
}

type UpdateSiteAddressInput struct {
	PinCode         *string `json:"pin_code"         binding:"omitempty,len=6,numeric"`
	City            *string `json:"city"             binding:"omitempty,min=1,max=100"`
	State           *string `json:"state"            binding:"omitempty,min=1,max=100"`
	AddressLine1    *string `json:"address_line_1"   binding:"omitempty,min=1,max=255"`
	AddressLine2    *string `json:"address_line_2"   binding:"omitempty,max=255"`
	SiteCoordinates *string `json:"site_coordinates" binding:"omitempty"`
}

type UpdateInsuranceInput struct {
	Price *float64 `json:"insurance_add_on_price" binding:"omitempty,gt=0"`
}

type UpdateTippingInput struct {
	// nil = keep the stored amounts. When sent: 1 to 10 values, each above 0.
	Amounts            []float64 `json:"tipping_amount"       binding:"omitempty,min=1,max=10,dive,gt=0"`
	PickTippingAllowed *bool     `json:"pick_tipping_allowed"`
	DropTippingAllowed *bool     `json:"drop_tipping_allowed"`
}

// IsEmpty reports whether the request changes nothing at all.
func (r *UpdateSiteRequest) IsEmpty() bool {
	return r.SiteName == nil &&
		r.IsBookingAllowed == nil &&
		r.ParkingType == nil &&
		r.TotalCapacity == nil &&
		r.DriverCount == nil &&
		r.OpeningTime == nil &&
		r.ClosingTime == nil &&
		r.ValetFee == nil &&
		r.SiteImages == nil &&
		r.SiteAddress == nil &&
		r.InsuranceAddOnAllowed == nil &&
		r.SiteInsuranceAddOn == nil &&
		r.TippingAllowed == nil &&
		r.TippingDetails == nil
}

// Validate checks rules that involve more than one field of the request.
// Rules that depend on the values already stored (for example turning
// insurance on when no price exists yet) are enforced by the database CHECK
func (r *UpdateSiteRequest) Validate() error {
	if r.IsEmpty() {
		return errors.New("at least one field must be provided to update")
	}

	if r.OpeningTime != nil && r.ClosingTime != nil && !r.OpeningTime.Before(*r.ClosingTime) {
		return errors.New("site_closing_time must be after site_opening_time")
	}

	if r.InsuranceAddOnAllowed != nil && !*r.InsuranceAddOnAllowed && r.SiteInsuranceAddOn != nil {
		return errors.New("site_insurance_add_on cannot be sent when insurance_add_on_allowed is false")
	}

	if r.TippingAllowed != nil && !*r.TippingAllowed && r.TippingDetails != nil {
		return errors.New("tipping_details cannot be sent when tipping_allowed is false")
	}

	return nil
}
