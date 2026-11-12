package dto

import (
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"time"
)

type CreateSlotRequest struct {
	SiteId    string `json:"site_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	IsOpen    bool   `json:"is_open" binding:"required"`
}

type UpdateSlotRequest struct {
	SlotId    string  `json:"slot_id" binding:"required"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	IsOpen    *bool   `json:"is_open"`
}

func (req *UpdateSlotRequest) ValidateUpdateSlotRequest() error {
	if req.StartTime == nil && req.EndTime == nil && req.IsOpen == nil {
		return CustomUtils.Invalid("At least one of start_time, end_time or is_open is required")
	} else if req.StartTime != nil && req.EndTime != nil {
		return ValidateTime(CustomUtils.Deref(req.StartTime), CustomUtils.Deref(req.EndTime))
	}
	return nil
}

func ValidateTime(s, e string) error {
	var err error
	var endTime, startTime time.Time

	startTime, err = time.Parse("15:04", s)
	endTime, err = time.Parse("15:04", e)
	if err != nil {
		return err
	}

	if !startTime.Before(endTime) {
		return CustomUtils.Invalid("Start time must be before end time")
	}
	return nil
}
