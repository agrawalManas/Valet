package service

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDriverRoleConflict = errors.New("This mobile number already belongs to a user who is not a driver")
	ErrDriverInactive     = errors.New("This driver account is deactivated")
)

// CreateDriver runs the whole flow in a single transaction:
//
//	1. check wether there exist a user with the given driver mobile number, if not then the below should be followed
//  2. find the supervisor by mobile number and check is_active (row locked)
//  3. add a new user in the users table with role DRIVER with relevant details (name, mobile, dob, gender, role)
//	4. create a new entry in driver_supervisors table for mapping
//  5. create a new entry in drivers table with other details pointing to site id
//
// If any step fails, nothing is written: no orphan user.

func AddDriverToSite(ctx context.Context, pool *pgxpool.Pool, supervisorId string, req dto.CreateDriverRequest) (*models.Driver, error) {
	dob, err := validateDriverRequest(req)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var driver *models.Driver

	// BeginFunc commits when the function returns nil and rolls back on any error or panic.
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		_, err := repository.GetUserByMobile(pool, req.MobileNumber)
		if err == nil {
			return invalid("A user with the given mobile number already exists, cannot add this driver")
		}

		s, err := getSupervisorDetails(pool, supervisorId)
		if err != nil {
			return err
		}

		d, err := resolveDriver(ctx, tx, req, dob)
		if err != nil {
			return err
		}

		_, err = repository.AddDriverSupervisorMapping(ctx, tx, d.Id, s.Id)
		if err != nil {
			return err
		}

		driver = buildDriverDetails(req, d.Id, s.Id)
		return repository.AddDriverToSite(ctx, tx, driver)
	})
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// validateDriverRequest covers rules the DTO tags and Validate() do not, and that
// the database CHECK constraints would otherwise reject with a 500.
// It returns the parsed date of birth.
func validateDriverRequest(req dto.CreateDriverRequest) (time.Time, error) {
	dob, err := time.Parse(constants.DOBLayout, req.Dob)
	if err != nil {
		return time.Time{}, invalid("driver_dob must be a valid date in YYYY-MM-DD format")
	}

	if dob.After(time.Now()) {
		return time.Time{}, invalid("driver_dob cannot be in the future")
	}

	//TODO: add age check

	if req.ShiftEndTime.Before(req.ShiftStartTime) {
		return time.Time{}, invalid("driver's shift end time cannot be after shift start time")
	}

	return dob, nil
}

// resolveDriver returns a locked, active user with the DRIVER role.
func resolveDriver(ctx context.Context, tx pgx.Tx, in dto.CreateDriverRequest, dob time.Time) (*models.User, error) {
	user, err := repository.CreateUserInTransaction(ctx, tx, in.FullName, in.Email, in.MobileNumber, constants.ROLE_DRIVER, in.Gender, dob)
	if err != nil {
		print("error", err.Error())
		return nil, err
	}

	if CustomUtils.Deref(user.Role) != constants.ROLE_DRIVER {
		return nil, ErrDriverRoleConflict
	}
	if !user.IsActive {
		return nil, ErrDriverInactive
	}

	return user, nil
}

func getSupervisorDetails(pool *pgxpool.Pool, supervisorId string) (*models.User, error) {
	user, err := repository.GetUserById(pool, supervisorId)
	if err != nil {
		return nil, err
	}

	if CustomUtils.Deref(user.Role) != constants.ROLE_SUPERVISOR {
		return nil, ErrSupervisorRoleConflict
	}
	if !user.IsActive {
		return nil, ErrSupervisorInactive
	}

	return user, nil
}

func buildDriverDetails(req dto.CreateDriverRequest, driverId string, supervisorId string) *models.Driver {
	driver := &models.Driver{
		Id:                driverId,
		SiteId:            req.SiteId,
		AddedBy:           supervisorId,
		ShiftStartTime:    req.ShiftStartTime,
		ShiftEndTime:      req.ShiftEndTime,
		AadhaarNumber:     req.AadhaarNumber,
		AadhaarUrl:        req.AadhaarUrl,
		PanNumber:         req.PanNumber,
		PanUrl:            req.PanUrl,
		DrivingLicense:    req.DrivingLicense,
		DrivingLicenseUrl: req.DrivingLicenseUrl,
		TotalExperience:   req.TotalExperience,
	}

	return driver
}

// Service function for changing the current supervisor for a specific driver in a single transaction
//
//  1. check wether there exist an active user with the given supervisor id, if found then lock the row
//  2. check whether there exists an active driver with given driver id, if found then lock the row
//  3. now update the supervisor id in driver_supervisor table
//  4. now update the current_supervisor in drivers table against this driver id
//
// If any step fails, nothing is written: no changes.
func ChangeSupervisorForDriver(ctx context.Context, pool *pgxpool.Pool, driverId, newSupervisorId string) (*models.Driver, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var driver *models.Driver
	var err error

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {

		d, err := repository.GetUserByIdAndLockTxn(ctx, tx, driverId, database.ForUpdate)
		if err != nil {
			return err
		}
		fmt.Printf("d role with type: %T", CustomUtils.Deref(d.Role))
		err = validateUser(constants.ROLE_DRIVER, nil, d)
		if err != nil {
			return err
		}

		s, err := repository.GetUserByIdAndLockTxn(ctx, tx, newSupervisorId, database.ForShare)
		if err != nil {
			return err
		}
		err = validateUser(constants.ROLE_SUPERVISOR, s, nil)
		if err != nil {
			return err
		}

		_, err = repository.UpdateDriverSupervisorMappingInTxn(ctx, tx, driverId, newSupervisorId)
		if err != nil {
			return err
		}

		driver, err = repository.UpdateCurrentDriverSupervisorInTxn(ctx, tx, driverId, newSupervisorId)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return driver, nil

}

func validateUser(checkFor string, supervisor *models.User, driver *models.User) error {

	switch checkFor {
	case constants.ROLE_SUPERVISOR:
		if CustomUtils.Deref(supervisor.Role) != constants.ROLE_SUPERVISOR {
			return invalid("The given supervisor id does not correspond to " + constants.ROLE_SUPERVISOR + " role")
		}
		if !supervisor.IsActive {
			return invalid("The given supervisor id is not an active user")
		}

	case constants.ROLE_DRIVER:
		if CustomUtils.Deref(driver.Role) != constants.ROLE_DRIVER {
			return invalid("The given driver id does not correspond to " + constants.ROLE_DRIVER + " role")
		}
		if !driver.IsActive {
			return invalid("The given driver id is not an active user")
		}
	}

	return nil
}
