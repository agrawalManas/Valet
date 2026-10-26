package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dobLayout = "2006-01-02"

var (
	ErrSupervisorRoleConflict = errors.New("this mobile number already belongs to a user who is not a supervisor")
	ErrSupervisorInactive     = errors.New("this supervisor account is deactivated")
)

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(msg string) error { return &ValidationError{Msg: msg} }

// CreateSite runs the whole flow in a single transaction:
//
//  1. find the supervisor by mobile number (row locked), or create the user
//  2. fill in any missing profile data (name, dob, gender) without overwriting
//  3. insert the site pointing at that user
//
// If any step fails, nothing is written: no orphan user, no half-created site.
// The caller must have bound the JSON and called req.Validate() already.

func CreateSite(ctx context.Context, pool *pgxpool.Pool, adminID string, req dto.AddSiteRequest) (*models.Site, error) {
	dob, err := validateSiteRequest(req)
	if err != nil {
		return nil, err
	}

	site := buildSite(req, adminID)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// BeginFunc commits when the function returns nil and rolls back on any error or panic.
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		supervisor, err := resolveSupervisor(ctx, tx, req.SupervisorDetails, dob)
		if err != nil {
			return err
		}

		site.SupervisorId = supervisor.Id
		site.SupervisorDetails = models.User{
			Id:           site.SupervisorId,
			FullName:     supervisor.FullName,
			MobileNumber: supervisor.MobileNumber,
			Dob:          supervisor.Dob,
			Gender:       supervisor.Gender,
		}
		return repository.InsertSite(ctx, tx, site)
	})
	if err != nil {
		return nil, err
	}
	return site, nil
}

// resolveSupervisor returns a locked, active user with the SUPERVISOR role for
// the given mobile number, creating the user if the number is new.
func resolveSupervisor(ctx context.Context, tx pgx.Tx, in dto.SupervisorInput, dob time.Time) (*models.User, error) {
	user, err := repository.FindUserByMobileForUpdate(ctx, tx, in.MobileNumber)

	switch {
	case err == nil:
		// existing user: validated below

	case errors.Is(err, pgx.ErrNoRows):
		user, err = repository.CreateUserInTransaction(ctx, tx, in.FullName, in.Email, in.MobileNumber, constants.ROLE_SUPERVISOR, in.Gender, dob)
		if errors.Is(err, pgx.ErrNoRows) {
			// A concurrent request created this number between our SELECT and INSERT.
			// The unique index made us wait for it to commit, so re-select now.
			user, err = repository.FindUserByMobileForUpdate(ctx, tx, in.MobileNumber)
		}
		if err != nil {
			return nil, err
		}

	default:
		return nil, err
	}

	if CustomUtils.Deref(user.Role) != constants.ROLE_SUPERVISOR {
		return nil, ErrSupervisorRoleConflict
	}
	if !user.IsActive {
		return nil, ErrSupervisorInactive
	}

	// Existing supervisors keep the data on file. Only gaps are filled from the request.
	if user.FullName == nil || user.Dob == nil || user.Gender == nil || user.Email == nil {
		if err := repository.FillUserProfileIfMissing(ctx, tx, user, in.FullName, in.Email, dob, in.Gender); err != nil {
			return nil, err
		}
	}
	return user, nil
}

// validateSiteRequest covers rules the DTO tags and Validate() do not, and that
// the database CHECK constraints would otherwise reject with a 500.
// It returns the parsed date of birth.
func validateSiteRequest(req dto.AddSiteRequest) (time.Time, error) {
	dob, err := time.Parse(dobLayout, req.SupervisorDetails.Dob)
	if err != nil {
		return time.Time{}, invalid("supervisor_dob must be a valid date in YYYY-MM-DD format")
	}
	if dob.After(time.Now()) {
		return time.Time{}, invalid("supervisor_dob cannot be in the future")
	}

	if CustomUtils.Deref(req.InsuranceAddOnAllowed) {
		if req.SiteInsuranceAddOn == nil || req.SiteInsuranceAddOn.Price == nil || *req.SiteInsuranceAddOn.Price <= 0 {
			return time.Time{}, invalid("insurance_add_on_price must be greater than 0 when the insurance add-on is allowed")
		}
	}

	if CustomUtils.Deref(req.TippingAllowed) {
		if req.TippingDetails == nil {
			return time.Time{}, invalid("tipping_details is required when tipping is allowed")
		}
		for i, amount := range req.TippingDetails.Amounts {
			if amount <= 0 {
				return time.Time{}, invalid(fmt.Sprintf("tipping_amount[%d] must be greater than 0", i))
			}
		}
	}

	return dob, nil
}

// buildSite maps the request onto the model. Server-owned fields (id, timestamps,
// supervisor_id) are filled in later. Insurance and tipping sections are kept only
// when their feature flag is on, so the stored data always matches the flags.
func buildSite(req dto.AddSiteRequest, adminID string) *models.Site {
	site := &models.Site{
		SiteName:         req.SiteName,
		IsBookingAllowed: CustomUtils.Deref(req.IsBookingAllowed),
		ParkingType:      req.ParkingType,
		TotalCapacity:    req.TotalCapacity,
		DriverCount:      req.DriverCount,
		OpeningTime:      req.OpeningTime,
		ClosingTime:      req.ClosingTime,
		ValetFee:         req.ValetFee,
		SiteImages:       req.SiteImages,
		SiteAddress: models.SiteAddress{
			PinCode:         req.SiteAddress.PinCode,
			City:            req.SiteAddress.City,
			State:           req.SiteAddress.State,
			AddressLine1:    req.SiteAddress.AddressLine1,
			AddressLine2:    req.SiteAddress.AddressLine2,
			SiteCoordinates: req.SiteAddress.SiteCoordinates,
		},
		InsuranceAddOnAllowed: CustomUtils.Deref(req.InsuranceAddOnAllowed),
		TippingAllowed:        CustomUtils.Deref(req.TippingAllowed),
		CreatedBy:             adminID,
	}

	if site.InsuranceAddOnAllowed && req.SiteInsuranceAddOn != nil {
		site.SiteInsuranceAddOn = &models.SiteInsuranceAddOn{
			InsuranceAddOnPrice: req.SiteInsuranceAddOn.Price,
		}
	}
	if site.TippingAllowed && req.TippingDetails != nil {
		site.TippingDetails = &models.TippingDetails{
			TippingAmount:      req.TippingDetails.Amounts,
			PickTippingAllowed: req.TippingDetails.PickTippingAllowed,
			DropTippingAllowed: req.TippingDetails.DropTippingAllowed,
		}
	}
	return site
}
