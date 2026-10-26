package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/models"
	"context"
)

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
