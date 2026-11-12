package models

import "time"

type Slot struct {
	Id        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SiteId    string    `json:"site_id"`
	IsOpen    bool      `json:"is_open"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
