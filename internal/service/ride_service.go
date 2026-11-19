package service

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//	CreateRideService runs upon the assumption that all the site/slot related validations were done while creating the booking,
//	allowing us to start the ride cycle. The following takes place in this transaction:
//
//	1. get booking details from booking id and check if the booking date is of today, the vehicle number matches and booking status is Scheduled
//	2. create a new ride entry with default ride status
//	3. create a new entry in ride history table and then add another entry if the current user is driver
//		as for the driver created bookings, the pickup_driver is caller only
//
// If any step fails, nothing is written: no orphan entry in the bookings table.
func CreateRideService(ctx context.Context, pool *pgxpool.Pool, bookingId, vrn string, currentUser middleware.AuthUser) (*models.Ride, *models.RideHistory, error) {
	var err error
	var ride *models.Ride
	var rideHistory *models.RideHistory

	ctx, cancel := context.WithTimeout(ctx, 500*time.Second)
	defer cancel()

	// TODO: add site capacity availability check

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var err error

		//---------Step 1
		bookingMode, err := resolveBooking(ctx, tx, bookingId, vrn)
		if err != nil {
			return err
		}

		//------------Step 2
		var driverId, supervisorId *string = nil, nil

		// for B2B bookings, set the supervisor and pickup_driver id beforehand
		if bookingMode == constants.BOOKING_MODE_B2B {
			switch currentUser.Role {
			case constants.ROLE_DRIVER:
				driverId = &currentUser.ID
			case constants.ROLE_SUPERVISOR:
				supervisorId = &currentUser.ID
			}
		}

		ride, err = repository.CreateNewRideInTxn(ctx, tx, bookingId, currentUser.ID, 8888, supervisorId, driverId)
		if err != nil {
			return err
		}

		//-------------Step 3
		req := buildRideHistoryReq(ride.Id, nil, constants.RIDE_STATUS_REQUESTED, &currentUser.ID, &currentUser.Role, nil)

		rideHistory, err = repository.AddRideHistoryInTxn(ctx, tx, req)
		if err != nil {
			return err
		}

		if driverId != nil {
			req := buildRideHistoryReq(ride.Id, &ride.RideStatus, constants.RIDE_STATUS_PICK_ASSIGNED, &currentUser.ID, &currentUser.Role, nil)
			rideHistory, err = repository.AddRideHistoryInTxn(ctx, tx, req)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return ride, rideHistory, nil
}

func resolveBooking(ctx context.Context, tx pgx.Tx, bookingId, vrn string) (string, error) {
	booking, err := repository.GetBookingByIdInTxn(ctx, tx, bookingId)
	if err != nil {
		return "", err
	}

	if vrn != booking.VehicleNumber {
		return "", CustomUtils.Invalid("This vehicle does not match with the one used when creating a booking")
	}

	if !CustomUtils.IsSameDate(booking.BookingDate, time.Now()) {
		return "", CustomUtils.Invalid("The booking date is not of today")
	}

	if booking.Status != constants.BOOKING_STATUS_SCHEDULED {
		return "", CustomUtils.Invalid("Cannot start a new ride for this booking as the current booking is " + booking.Status)
	}
	return booking.BookingMode, nil
}

func buildRideHistoryReq(rideId string, fromStatus *string, toStatus string, actorId, actorRole, driverId *string) *models.RideHistory {
	rideHistory := &models.RideHistory{
		RideId:         rideId,
		PreviousStatus: fromStatus,
		NewStatus:      toStatus,
		ActorId:        actorId,
		ActorRole:      actorRole,
		DriverId:       driverId,
	}

	return rideHistory
}
