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
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateBooking runs the whole flow in a single transaction:
//  1. get the site details using the site id and check if bookings are allowed
//  2. check the total bookings for this site against the capacity of the site, also if there exist a current active/scheduled booking for the given vehicle number
//  3. get the slot details using the given slot id and verify if this slot is associated with given site id and also compare the booking time
//  4. check wether there exist an active user with the given mobile number, if not then create a new user and retrieve the user id
//  5. compute the taxes and total amount(valet + insurance + taxes)
//  6. create a new entry in bookings table
//
// If any step fails, nothing is written: no orphan entry in the bookings table.
func CreateBooking(ctx context.Context, pool *pgxpool.Pool, createdBy string, req dto.CreateBookingRequest) (*models.Booking, error) {
	var err error
	ctx, cancel := context.WithTimeout(ctx, 500*time.Second)
	defer cancel()

	var booking *models.Booking

	// BeginFunc commits when the function returns nil and rolls back on any error or panic.
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var insuranceAmount float64 = 0

		//--------------Step 1
		site, err := repository.GetSiteBySiteId(pool, req.SiteId)
		if err != nil {
			return err
		}

		if !site.IsBookingAllowed {
			return CustomUtils.Invalid("The bookings for site are disabled at the moment")
		}

		if !site.InsuranceAddOnAllowed && req.InsuranceOpted {
			return CustomUtils.Invalid("This site does not allow insurance, please opt out")
		} else if site.InsuranceAddOnAllowed && req.InsuranceOpted {
			if site.SiteInsuranceAddOn.InsuranceAddOnPrice != nil {
				insuranceAmount = CustomUtils.Deref(site.SiteInsuranceAddOn.InsuranceAddOnPrice)
			}
		}

		//--------------Step 2
		currentBookingCount, err := repository.GetBookingCountForASiteInTransaction(ctx, tx, req.SiteId, req.BookingDate)
		if err != nil {
			return err
		}
		if CustomUtils.Deref(currentBookingCount) > site.TotalCapacity {
			return CustomUtils.Invalid("Not enough remaining spots for a new booking")
		}

		currentBookingCount, err = repository.GetActiveScheduledBookingsForVRN(ctx, tx, req.SiteId, req.VehicleNumber)
		if err != nil {
			return err
		}
		if CustomUtils.Deref(currentBookingCount) == 1 {
			return CustomUtils.Invalid("This vehicle has an active or scheduled booking, cannot create a new one")
		}

		//--------------Step 3
		slot, err := repository.GetSlotByIdAndLockTxn(ctx, tx, req.SlotId, req.SiteId, database.ForShare)
		if err != nil {
			return err
		}

		if !slot.IsOpen {
			return CustomUtils.Invalid("This slot is not active at the moment")
		}

		//---------------Step 4
		user, err := repository.FindUserByMobileForUpdate(ctx, tx, req.MobileNumber)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				user, err = repository.CreateUserInTransaction(ctx, tx, nil, nil, req.MobileNumber, constants.ROLE_CUSTOMER, nil, nil)
				if err != nil {
					return CustomUtils.Invalid(err.Error())
				}
				return err
			}
		}

		if !user.IsActive {
			return CustomUtils.Invalid("Cannot create a booking for an inactive user")
		}

		//-----------------Step 5
		var cgst float64 = 0.09
		var sgst float64 = 0.09
		var totalTaxes float64 = (site.ValetFee * cgst) + (site.ValetFee * sgst)

		//------------------Step 6
		bookingReq := buildBooking(createdBy, site.ValetFee, totalTaxes, insuranceAmount, req)
		var paymentStatus string

		if req.PaymentStatus == nil {
			paymentStatus = "PENDING"
		} else {
			paymentStatus = CustomUtils.Deref(req.PaymentStatus)
		}
		booking, err = repository.CreateBookingInTxn(ctx, tx, user.Id, createdBy, paymentStatus, bookingReq)
		return err
	})
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func buildBooking(createdBy string, valetAmount, totalTax, insuranceAmount float64, req dto.CreateBookingRequest) *models.Booking {
	var booking *models.Booking

	booking = &models.Booking{
		CreatedBy:       createdBy,
		SiteId:          req.SiteId,
		SlotId:          req.SlotId,
		VehicleNumber:   req.VehicleNumber,
		VehicleName:     req.VehicleName,
		BookingDate:     req.BookingDate,
		BookingType:     req.BookingType,
		BookingMode:     req.BookingMode,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   req.PaymentStatus,
		ValetAmount:     valetAmount,
		TotalTax:        totalTax,
		InsuranceOpted:  req.InsuranceOpted,
		InsuranceAmount: insuranceAmount,
	}
	return booking
}
