package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const siteColumns = `site_name, is_booking_allowed, parking_type, total_capacity, driver_count,
			site_opening_time, site_closing_time, valet_fee, site_images,
			pin_code, city, state, address_line_1, address_line_2, site_coordinates,
			supervisor_id,
			insurance_add_on_allowed, insurance_add_on_price,
			tipping_allowed, tipping_amount, pick_tipping_allowed, drop_tipping_allowed,
			created_by`

func scanSite(row pgx.Row) (*models.Site, error) {
	print("in the scanSite method")
	var s models.Site
	var insurancePrice *float64
	var tippingAmounts []float64
	var pickTipping, dropTipping *bool

	err := row.Scan(
		&s.Id, &s.SiteName, &s.IsBookingAllowed, &s.ParkingType, &s.TotalCapacity, &s.DriverCount,
		&s.OpeningTime, &s.ClosingTime, &s.ValetFee, &s.SiteImages,
		&s.SiteAddress.PinCode, &s.SiteAddress.City, &s.SiteAddress.State,
		&s.SiteAddress.AddressLine1, &s.SiteAddress.AddressLine2, &s.SiteAddress.SiteCoordinates,
		&s.SupervisorId,
		&s.InsuranceAddOnAllowed, &insurancePrice,
		&s.TippingAllowed, &tippingAmounts, &pickTipping, &dropTipping,
		&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		print("error is: ", err.Error())
		return nil, err
	}

	if s.InsuranceAddOnAllowed {
		s.SiteInsuranceAddOn = &models.SiteInsuranceAddOn{InsuranceAddOnPrice: insurancePrice}
	}
	if s.TippingAllowed {
		s.TippingDetails = &models.TippingDetails{
			TippingAmount:      tippingAmounts,
			PickTippingAllowed: pickTipping,
			DropTippingAllowed: dropTipping,
		}
	}
	return &s, nil
}

func InsertSite(ctx context.Context, db database.DBTX, s *models.Site) error {
	var (
		insurancePrice *float64
		tippingAmounts []float64 // nil is stored as NULL
		pickTipping    *bool
		dropTipping    *bool
	)
	if s.SiteInsuranceAddOn != nil {
		insurancePrice = s.SiteInsuranceAddOn.InsuranceAddOnPrice
	}
	if s.TippingDetails != nil {
		tippingAmounts = s.TippingDetails.TippingAmount
		pickTipping = s.TippingDetails.PickTippingAllowed
		dropTipping = s.TippingDetails.DropTippingAllowed
	}

	images := s.SiteImages
	if images == nil {
		images = []string{} // site_images is NOT NULL
	}

	const query = `
		INSERT INTO sites (
			site_name, is_booking_allowed, parking_type, total_capacity, driver_count,
			site_opening_time, site_closing_time, valet_fee, site_images,
			pin_code, city, state, address_line_1, address_line_2, site_coordinates,
			supervisor_id,
			insurance_add_on_allowed, insurance_add_on_price,
			tipping_allowed, tipping_amount, pick_tipping_allowed, drop_tipping_allowed,
			created_by
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16,
			$17, $18,
			$19, $20, $21, $22,
			$23
		)
		RETURNING id, supervisor_id, created_at, updated_at`

	return db.QueryRow(ctx, query,
		s.SiteName, s.IsBookingAllowed, s.ParkingType, s.TotalCapacity, s.DriverCount,
		s.OpeningTime, s.ClosingTime, s.ValetFee, images,
		s.SiteAddress.PinCode, s.SiteAddress.City, s.SiteAddress.State,
		s.SiteAddress.AddressLine1, s.SiteAddress.AddressLine2, s.SiteAddress.SiteCoordinates,
		s.SupervisorId,
		s.InsuranceAddOnAllowed, insurancePrice,
		s.TippingAllowed, tippingAmounts, pickTipping, dropTipping,
		s.CreatedBy,
	).Scan(&s.Id, &s.SupervisorId, &s.CreatedAt, &s.UpdatedAt)
}

func GetSites(pool *pgxpool.Pool, createdBy *string, supervisorId *string) ([]models.Site, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sites []models.Site = []models.Site{}
	var query string
	var rows pgx.Rows
	var err error

	if createdBy != nil {
		query = `SELECT s.id, ` + siteColumns + `, s.created_at, s.updated_at,
				 u.id, u.full_name, u.email, u.mobile_number, u.is_active,
				 u.gender, u.dob::text, u.created_at, u.updated_at
				 FROM sites s JOIN users u ON u.id = s.created_by
				 WHERE created_by = $1
				 ORDER BY s.created_at DESC`

		rows, err = pool.Query(ctx, query, createdBy)
	} else if supervisorId != nil {
		query = `SELECT s.id, ` + siteColumns + `, s.created_at, s.updated_at, 
				 u.id, u.full_name, u.email, u.mobile_number, u.is_active,
				 u.gender, u.dob::text, u.created_at, u.updated_at
				 FROM sites s JOIN users u ON u.id = s.supervisor_id
				 WHERE s.supervisor_id = $1
				 ORDER BY s.created_at DESC`

		rows, err = pool.Query(ctx, query, supervisorId)
	} else {
		query = `SELECT s.id, ` + siteColumns + `, s.created_at, s.updated_at,
				 u.id, u.full_name, u.email, u.mobile_number, u.is_active,
       			 u.gender, u.dob::text, u.created_at, u.updated_at
	   			 FROM sites s JOIN users u ON u.id = s.supervisor_id`
		rows, err = pool.Query(ctx, query)
	}

	if err != nil || rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Site
		var insurancePrice *float64
		var tippingAmounts []float64
		var pickTipping, dropTipping *bool

		err := rows.Scan(
			&s.Id, &s.SiteName, &s.IsBookingAllowed, &s.ParkingType, &s.TotalCapacity, &s.DriverCount,
			&s.OpeningTime, &s.ClosingTime, &s.ValetFee, &s.SiteImages,
			&s.SiteAddress.PinCode, &s.SiteAddress.City, &s.SiteAddress.State,
			&s.SiteAddress.AddressLine1, &s.SiteAddress.AddressLine2, &s.SiteAddress.SiteCoordinates,
			&s.SupervisorId,
			&s.InsuranceAddOnAllowed, &insurancePrice,
			&s.TippingAllowed, &tippingAmounts, &pickTipping, &dropTipping,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
			&s.SupervisorDetails.Id, &s.SupervisorDetails.FullName, &s.SupervisorDetails.Email, &s.SupervisorDetails.MobileNumber,
			&s.SupervisorDetails.IsActive, &s.SupervisorDetails.Gender, &s.SupervisorDetails.Dob, &s.SupervisorDetails.CreatedAt, &s.SupervisorDetails.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if s.InsuranceAddOnAllowed {
			s.SiteInsuranceAddOn = &models.SiteInsuranceAddOn{InsuranceAddOnPrice: insurancePrice}
		}
		if s.TippingAllowed {
			s.TippingDetails = &models.TippingDetails{
				TippingAmount:      tippingAmounts,
				PickTippingAllowed: pickTipping,
				DropTippingAllowed: dropTipping,
			}
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func UpdateSite(pool *pgxpool.Pool, siteId string, newDetails dto.UpdateSiteRequest, supervisorId string) (*models.Site, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pinCode, city, state, addressLine1, addressLine2, siteCoordinates *string

	if newDetails.SiteAddress != nil {
		pinCode = newDetails.SiteAddress.PinCode
		city = newDetails.SiteAddress.City
		state = newDetails.SiteAddress.State
		addressLine1 = newDetails.SiteAddress.AddressLine1
		addressLine2 = newDetails.SiteAddress.AddressLine2
		siteCoordinates = newDetails.SiteAddress.SiteCoordinates
	}

	// Insurance: nil means keep the existing value.
	var insurancePrice *float64

	if newDetails.SiteInsuranceAddOn != nil {
		insurancePrice = newDetails.SiteInsuranceAddOn.Price
	}

	// Tipping: nil means keep the existing values.
	var tippingAmounts []float64
	var pickTippingAllowed, dropTippingAllowed *bool

	if newDetails.TippingDetails != nil {
		tippingAmounts = newDetails.TippingDetails.Amounts
		pickTippingAllowed = newDetails.TippingDetails.PickTippingAllowed
		dropTippingAllowed = newDetails.TippingDetails.DropTippingAllowed
	}

	query := ` UPDATE sites
			   SET site_name = COALESCE($2, site_name),
			   is_booking_allowed = COALESCE($3, is_booking_allowed),
			   parking_type = COALESCE($4, parking_type),
			   total_capacity = COALESCE($5, total_capacity),
			   driver_count = COALESCE($6, driver_count),
			   site_opening_time = COALESCE($7, site_opening_time),
			   site_closing_time = COALESCE($8, site_closing_time),
			   valet_fee = COALESCE($9, valet_fee),
			   site_images = COALESCE($10, site_images),
			   pin_code = COALESCE($11, pin_code),
			   city = COALESCE($12, city),
			   state = COALESCE($13, state),
			   address_line_1 = COALESCE($14, address_line_1),
			   address_line_2 = COALESCE($15, address_line_2),
			   site_coordinates = COALESCE($16, site_coordinates),
			   supervisor_id = COALESCE($17, supervisor_id),
			   insurance_add_on_allowed = COALESCE($18, insurance_add_on_allowed),
			   insurance_add_on_price = COALESCE($19, insurance_add_on_price),
			   tipping_allowed = COALESCE($20, tipping_allowed),
			   tipping_amount = COALESCE($21, tipping_amount),
			   pick_tipping_allowed = COALESCE($22, pick_tipping_allowed),
			   drop_tipping_allowed = COALESCE($23, drop_tipping_allowed),
			   updated_at = CURRENT_TIMESTAMP
			   WHERE id = $1
			   RETURNING id, ` + siteColumns + `, created_at, updated_at`

	row := pool.QueryRow(ctx, query, siteId, newDetails.SiteName, newDetails.IsBookingAllowed, newDetails.ParkingType, newDetails.TotalCapacity,
		newDetails.DriverCount, newDetails.OpeningTime, newDetails.ClosingTime, newDetails.ValetFee,
		newDetails.SiteImages, pinCode, city, state, addressLine1, addressLine2, siteCoordinates, supervisorId, newDetails.InsuranceAddOnAllowed,
		insurancePrice, newDetails.TippingAllowed, tippingAmounts, pickTippingAllowed, dropTippingAllowed)

	return scanSite(row)
}
