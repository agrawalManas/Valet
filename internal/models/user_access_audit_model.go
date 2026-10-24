package models

import "time"

type UserAccessAudit struct {
	Id          int       `json:"id"`
	UserId      string    `json:"user_id"`
	ChangedBy   string    `json:"changed_by"`
	OldRole     *string   `json:"old_role"`
	NewRole     *string   `json:"new_role"`
	OldIsActive *bool     `json:"old_is_active"`
	NewIsActive *bool     `json:"new_is_active"`
	Reason      *string   `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}
