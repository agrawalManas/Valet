package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/models"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const driverColumns = `id, site_id, created_by, shift_start_time, shift_end_time, current_status, last_login, last_logout, total_tip_earned,
aadhaar_number, aadhaar_url, pan_number, pan_url, driving_license, driving_license_url, kyc_status, driving_experience_years, created_at, updated_at, current_supervisor`

var driverJoinColumns string = CustomUtils.PrefixColumns("d", driverColumns)

func scanDriver(row pgx.Row, joinDriverDetails bool) (*models.Driver, error) {
	var driver models.Driver
	var err error

	if joinDriverDetails {
		var driverInfo models.User
		err = row.Scan(
			&driver.Id, &driver.SiteId, &driver.AddedBy, &driver.ShiftStartTime, &driver.ShiftEndTime, &driver.CurrentStatus,
			&driver.LastLogin, &driver.LastLogout, &driver.TotalTipEarned, &driver.AadhaarNumber,
			&driver.AadhaarUrl, &driver.PanNumber, &driver.PanUrl,
			&driver.DrivingLicense, &driver.DrivingLicenseUrl, &driver.KYCStatus,
			&driver.TotalExperience, &driver.CreatedAt, &driver.UpdatedAt, &driver.CurrentSupervisor, &driverInfo.Id, &driverInfo.FullName, &driverInfo.Email, &driverInfo.MobileNumber,
			&driverInfo.Role, &driverInfo.IsActive, &driverInfo.Dob, &driverInfo.Gender, &driverInfo.CreatedAt, &driverInfo.UpdatedAt)

		driver.DriverDetails = &driverInfo

	} else {
		err = row.Scan(
			&driver.Id, &driver.SiteId, &driver.AddedBy, &driver.ShiftStartTime, &driver.ShiftEndTime, &driver.CurrentStatus,
			&driver.LastLogin, &driver.LastLogout, &driver.TotalTipEarned, &driver.AadhaarNumber,
			&driver.AadhaarUrl, &driver.PanNumber, &driver.PanUrl,
			&driver.DrivingLicense, &driver.DrivingLicenseUrl, &driver.KYCStatus,
			&driver.TotalExperience, &driver.CreatedAt, &driver.UpdatedAt, &driver.CurrentSupervisor)
	}
	if err != nil {
		return nil, err
	}

	return &driver, nil
}

func AddDriverToSite(ctx context.Context, db database.DBTX, driver *models.Driver) error {

	const query = `
		INSERT INTO drivers (
			id, site_id, created_by, shift_start_time, shift_end_time, aadhaar_number,
			aadhaar_url, pan_number, pan_url, driving_license, driving_license_url, driving_experience_years, current_supervisor
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING ` + driverColumns

	return db.QueryRow(ctx, query, driver.Id, driver.SiteId, driver.AddedBy, driver.ShiftStartTime, driver.ShiftEndTime,
		driver.AadhaarNumber, driver.AadhaarUrl, driver.PanNumber, driver.PanUrl, driver.DrivingLicense, driver.DrivingLicenseUrl,
		driver.TotalExperience, driver.CurrentSupervisor).Scan(&driver.Id, &driver.SiteId, &driver.AddedBy, &driver.ShiftStartTime, &driver.ShiftEndTime,
		&driver.CurrentStatus, &driver.LastLogin, &driver.LastLogout, &driver.TotalTipEarned, &driver.AadhaarNumber, &driver.AadhaarUrl,
		&driver.PanNumber, &driver.PanUrl, &driver.DrivingLicense, &driver.DrivingLicenseUrl, &driver.KYCStatus, &driver.TotalExperience,
		&driver.CreatedAt, &driver.UpdatedAt, &driver.CurrentSupervisor)
}

func AddDriverSupervisorMapping(ctx context.Context, db database.DBTX, driverId string, supervisorId string) (*models.DriverSupervisors, error) {
	var driverSupervisor models.DriverSupervisors

	const query = `
		INSERT INTO driver_supervisors (driver_id, supervisor_id)
		VALUES ($1, $2)
		RETURNING driver_id, supervisor_id, assigned_at`

	err := db.QueryRow(ctx, query, driverId, supervisorId).Scan(&driverSupervisor.DriverId, &driverSupervisor.SupervisorId, &driverSupervisor.AssignedAt)
	if err != nil {
		return nil, err
	}
	return &driverSupervisor, nil
}

func GetAllDriversForSite(pool *pgxpool.Pool, id string, bySiteId bool) ([]*models.Driver, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var drivers []*models.Driver = []*models.Driver{}
	var query string
	var rows pgx.Rows
	var err error

	if bySiteId {
		query = `SELECT ` + driverJoinColumns + ", " + userJoinColumns + `
			 FROM drivers d
			 JOIN users u ON u.id = d.id
			 WHERE site_id = $1
			 ORDER BY u.created_at DESC`
	} else {
		query = `SELECT ` + driverJoinColumns + ", " + userJoinColumns + `
			 FROM drivers d
			 JOIN users u ON u.id = d.id
			 WHERE created_by = $1
			 ORDER BY u.created_at DESC`
	}

	rows, err = pool.Query(ctx, query, id)
	if err != nil || rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		driver, err := scanDriver(rows, true)
		if err != nil {
			return nil, err
		}
		drivers = append(drivers, driver)
	}
	return drivers, rows.Err()
}

func UpdateDriver(pool *pgxpool.Pool, driverId string, currentStatus *string, shiftStartTime, shiftEndTime *time.Time, experience *int16, supervisorId string) (*models.Driver, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE drivers 
			  SET current_status = COALESCE($2, current_status),
			  shift_start_time = COALESCE($3, shift_start_time),
			  shift_end_time = COALESCE($4, shift_end_time),
			  driving_experience_years = COALESCE($5, driving_experience_years),
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1 AND current_supervisor = $6
			  RETURNING ` + driverColumns

	row := pool.QueryRow(ctx, query, driverId, currentStatus, shiftStartTime, shiftEndTime, experience, supervisorId)
	return scanDriver(row, false)
}

func UpdateDriverSupervisorMappingInTxn(ctx context.Context, db database.DBTX, driverId, supervisorId string) (*models.DriverSupervisors, error) {
	var driverSupervisor models.DriverSupervisors

	const query = `UPDATE driver_supervisors
		SET supervisor_id = $2
		WHERE driver_id = $1
		RETURNING driver_id, supervisor_id, assigned_at`

	err := db.QueryRow(ctx, query, driverId, supervisorId).Scan(&driverSupervisor.DriverId, &driverSupervisor.SupervisorId, &driverSupervisor.AssignedAt)
	if err != nil {
		return nil, err
	}
	return &driverSupervisor, nil
}

func UpdateCurrentDriverSupervisorInTxn(ctx context.Context, db database.DBTX, driverId, supervisorId string) (*models.Driver, error) {
	const query = `UPDATE drivers
		SET current_supervisor = $2
		WHERE id = $1
		RETURNING ` + driverColumns

	row := db.QueryRow(ctx, query, driverId, supervisorId)
	return scanDriver(row, false)
}
