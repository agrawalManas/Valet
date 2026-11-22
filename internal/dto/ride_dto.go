package dto

import (
	"agrawalManas/Valet/internal/constants"
	CustomUtils "agrawalManas/Valet/pkg/utils"
)

type UpdateRideRequest struct {
	RideID          string   `json:"ride_id" binding:"required"`
	BookingId       string   `json:"booking_id" binding:"required"`
	PreviousStatus  string   `json:"previous_status" binding:"required,oneof=REQUESTED PICK_ASSIGNED PICK_COMPLETED PARKED GMC_REQUESTED DROP_ASSIGNED DROP_ENROUTE DROP_COMPLETED KEY_RETURNED RIDE_FEEDBACK"`
	SupervisorId    *string  `json:"supervisor_id" binding:"omitempty"` // optional for driver, as key can be returned to supervisor or left at the desk
	RideReviewId    *string  `json:"ride_review_id" binding:"omitempty"`
	ParkingLocation *string  `json:"parking_location" binding:"omitempty"`
	ParkedCarImages []string `json:"parked_car_images" binding:"omitempty,max=5,dive,url"`
}

type fieldRule struct {
	required []string
	optional []string
}

var updateRideFields = []string{
	"supervisor_id", "ride_review_id", "parking_location", "parked_car_images",
}

var updateRideRules = map[string]fieldRule{
	constants.RIDE_STATUS_REQUESTED:      {},
	constants.RIDE_STATUS_PICK_ASSIGNED:  {},
	constants.RIDE_STATUS_PICK_COMPLETED: {optional: []string{"parking_location", "parked_car_images"}},
	constants.RIDE_STATUS_PARKED:         {},
	constants.RIDE_STATUS_GMC_REQUESTED:  {},
	constants.RIDE_STATUS_DROP_ASSIGNED:  {},
	constants.RIDE_STATUS_DROP_ENROUTE:   {},
	constants.RIDE_STATUS_DROP_COMPLETED: {optional: []string{"supervisor_id"}},
	constants.RIDE_STATUS_KEY_RETURNED:   {},
	constants.RIDE_STATUS_FEEDBACK:       {optional: []string{"ride_review_id"}},
}

func (r *UpdateRideRequest) sentFields() map[string]bool {
	return map[string]bool{
		"supervisor_id":     r.SupervisorId != nil,
		"ride_review_id":    r.RideReviewId != nil,
		"parking_location":  r.ParkingLocation != nil,
		"parked_car_images": len(r.ParkedCarImages) > 0,
	}
}

func (r *UpdateRideRequest) Validate(requestGMCAllowed bool) error {
	if r.PreviousStatus == constants.RIDE_STATUS_PARKED && !requestGMCAllowed {
		return CustomUtils.Invalid("Please ask to customer/supervisor to request GMC")
	}
	rule, ok := updateRideRules[r.PreviousStatus]
	if !ok {
		return CustomUtils.Invalid("a ride with status " + r.PreviousStatus + " cannot be updated")
	}

	sent := r.sentFields()
	allowed := make(map[string]bool)

	for _, f := range rule.required {
		allowed[f] = true
		if !sent[f] {
			return CustomUtils.Invalid(f + " is required when the previous_status is " + r.PreviousStatus)
		}
	}
	for _, f := range rule.optional {
		allowed[f] = true
	}

	for _, f := range updateRideFields {
		if sent[f] && !allowed[f] {
			return CustomUtils.Invalid(f + " is not allowed when previous_status is " + r.PreviousStatus)
		}
	}
	return nil
}
