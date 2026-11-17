package dto

import (
	"agrawalManas/Valet/internal/constants"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"time"
)

type CreateBookingRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required,max=10"`
	SiteId       string `json:"site_id" binding:"required"`
	SlotId       string `json:"slot_id" binding:"required"`

	VehicleNumber string `json:"vehicle_number" binding:"required"`
	VehicleName   string `json:"vehicle_name" binding:"required"`

	BookingDate time.Time `json:"booking_date" binding:"required"`
	BookingType string    `json:"booking_type" binding:"required,oneof=PREPAID GMC"`
	BookingMode string    `json:"booking_mode" binding:"required,oneof=B2B B2C"`

	InsuranceOpted bool `json:"insurance_opted" binding:"required"`

	PaymentMethod *string `json:"payment_method" binding:"omitempty,oneof=CASH UPI CARD NET_BANKING"`
	PaymentStatus *string `json:"payment_status" binding:"omitempty,oneof=NOT_REQUIRED PENDING PAID FAILED REFUNDED"`
}

func (req *CreateBookingRequest) ValidateBookingRequest() error {
	var err error
	// validate mobile_number
	err = CustomUtils.ValidateMobileNumbers(req.MobileNumber)
	if err != nil {
		return err
	}

	// validate if all the ids are correct uuids
	err = CustomUtils.ValidateUUIDs(req.SiteId, req.SlotId)
	if err != nil {
		return err
	}

	// validate vehicle number
	err = CustomUtils.ValidateVRNs(req.VehicleNumber)
	if err != nil {
		return err
	}

	// validate booking date; should not be before today
	if CustomUtils.IsDateBefore(req.BookingDate, time.Now()) {
		return CustomUtils.Invalid("Booking date should be today or future")
	}

	// check the booking type and payment info
	if req.BookingType == constants.BOOKING_TYPE_GMC {
		if req.PaymentStatus != nil && req.PaymentMethod != nil {
			return CustomUtils.Invalid("Cannot add payment details with Booking Type = GMC")
		}
	} else {
		if req.PaymentMethod == nil {
			return CustomUtils.Invalid("Payment Method is required")
		}
		if CustomUtils.Deref(req.PaymentStatus) != constants.PAYMENT_STATUS_PAID {
			return CustomUtils.Invalid("Payment Status not valid for pre-booking")
		}
	}

	return nil
}
