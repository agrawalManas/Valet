package models

import "time"

type DriverSupervisors struct {
	DriverId     string    `json:"driver_id"`
	SupervisorId string    `json:"supervisor_id"`
	AssignedAt   time.Time `json:"assigned_at"`
}
