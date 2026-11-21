package service

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/dto"
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
//	2. for the B2B requests check if it is from Driver role, if yes then make sure if this driver is linked to given site id,
//	   if verified then set him as pickup_driver
//	3. update booking status to active in bookings table against the booking id
//	4. create a new ride entry with default ride status
//	5. create a new entry in ride history table and then add another entry if the current user is driver
//		as for the driver created rides, the pickup_driver is caller only
//
// If any step fails, nothing is written: no orphan entry in the rides and ride_history table.
func CreateRideService(ctx context.Context, pool *pgxpool.Pool, bookingId, vrn string, currentUser middleware.AuthUser) (*models.Ride, error) {
	var err error
	var ride *models.Ride

	ctx, cancel := context.WithTimeout(ctx, 500*time.Second)
	defer cancel()

	// TODO: add site capacity availability check

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var err error

		//---------Step 1
		booking, err := resolveBooking(ctx, tx, bookingId, vrn)
		if err != nil {
			return err
		}

		//------------Step 2
		var driverId, supervisorId *string = nil, nil
		siteId := booking.SiteId

		if booking.BookingMode == constants.BOOKING_MODE_B2B {
			driverId, supervisorId, err = resolveB2BBookingMode(ctx, tx, currentUser.ID, currentUser.Role, siteId)
			if err != nil {
				return err
			}
		}

		// -----------Step 3
		_, err = repository.UpdateBookingAndPaymentStatusInTxn(ctx, tx, bookingId, CustomUtils.GetPtr(constants.BOOKING_STATUS_ACTIVE), nil)
		if err != nil {
			return err
		}

		//------------Step 4
		ride, err = repository.CreateNewRideInTxn(ctx, tx, bookingId, siteId, currentUser.ID, 8888, supervisorId, driverId)
		if err != nil {
			return err
		}

		//-------------Step 5
		req := buildRideHistoryReq(ride.Id, nil, constants.RIDE_STATUS_REQUESTED, &currentUser.ID, &currentUser.Role, nil)

		_, err = repository.AddRideHistoryInTxn(ctx, tx, req)
		if err != nil {
			return err
		}

		if driverId != nil {
			req := buildRideHistoryReq(ride.Id, &ride.RideStatus, constants.RIDE_STATUS_PICK_ASSIGNED, &currentUser.ID, &currentUser.Role, &currentUser.ID)
			_, err = repository.AddRideHistoryInTxn(ctx, tx, req)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ride, nil
}

func resolveB2BBookingMode(ctx context.Context, tx database.DBTX, currentUserId, currentUserRole, siteId string) (driverId, supervisorId *string, err error) {
	switch currentUserRole {
	case constants.ROLE_DRIVER:
		var exists bool
		exists, err = repository.CheckDriverAgainstSiteIdInTx(ctx, tx, currentUserId, siteId)
		if err != nil {
			return nil, nil, err
		}

		if !exists {
			return nil, nil, CustomUtils.Invalid("This driver doest not belong to the given site id")
		}
		driverId = &currentUserId
	case constants.ROLE_SUPERVISOR:
		supervisorId = &currentUserId
	}
	return driverId, supervisorId, err
}

func resolveBooking(ctx context.Context, tx database.DBTX, bookingId, vrn string) (*models.Booking, error) {
	booking, err := repository.GetBookingByIdInTxn(ctx, tx, bookingId, database.ForUpdate)
	if err != nil {
		return nil, err
	}

	if vrn != booking.VehicleNumber {
		return nil, CustomUtils.Invalid("This vehicle does not match with the one used when creating a booking")
	}

	if !CustomUtils.IsSameDate(booking.BookingDate, time.Now()) {
		return nil, CustomUtils.Invalid("The booking date is not of today")
	}

	if booking.Status != constants.BOOKING_STATUS_SCHEDULED {
		return nil, CustomUtils.Invalid("Cannot start a new ride for this booking as the current booking is " + booking.Status)
	}
	return booking, nil
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

//	UpdateRideService runs upon the assumption that all the site/slot related validations were done while creating the booking,
//	allowing us to start the ride cycle. The following takes place in this transaction:
//
//	1. get ride details from ride id and check the current ride status is same previous_status
//	2. if it is driver request, check whether the driver is associated with site or not then update the current entry
//	3. create a new ride entry with the new ride statuses
//	4. update booking status to complete if applicable
//
// If any step fails, nothing is written: no orphan entry in the rides and ride_history table.
func UpdateRideService(ctx context.Context, pool *pgxpool.Pool, req dto.UpdateRideRequest, currentUser middleware.AuthUser) (*models.Ride, error) {
	var err error
	var updatedRide *models.Ride

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	previousStatus := req.PreviousStatus
	isPickOrDropRequest := previousStatus == constants.RIDE_STATUS_REQUESTED || previousStatus == constants.RIDE_STATUS_GMC_REQUESTED
	if isPickOrDropRequest && currentUser.Role != constants.ROLE_DRIVER {
		err = CustomUtils.Invalid("This ride can be moved forward when a driver accepts the request")
	}
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// -----------Step 1
		ride, err := repository.GetRideDetailsByID(ctx, tx, req.RideID, database.ForUpdate)
		if err != nil {
			return err
		}

		var status = ride.RideStatus
		if previousStatus != status {
			return CustomUtils.Invalid("Current ride status " + status + " does not match the given previous_status")
		}

		// -----------Step 2
		if isPickOrDropRequest {
			validDriver, err := repository.CheckDriverAgainstSiteIdInTx(ctx, tx, currentUser.ID, ride.SiteId)
			if err != nil {
				return err
			}

			if !validDriver {
				return CustomUtils.Invalid("Driver not associated with this site")
			}
		}
		updatedRide, err = repository.StepUpRideInTxn(ctx, tx, req)
		if err != nil {
			// if errors.Is(err, pgx.ErrNoRows) {
			// 	return CustomUtils.Invalid("No ride found with given ride_id")
			// }
			return err
		}

		// ------------Step 3
		r := buildRideHistoryReq(req.RideID, &req.PreviousStatus, updatedRide.RideStatus, &currentUser.ID, &currentUser.Role, getDriverIdForRideHistory(req))
		_, err = repository.AddRideHistoryInTxn(ctx, tx, r)
		if err != nil {
			return err
		}

		// --------------Step 4
		if updatedRide.RideStatus == constants.RIDE_STATUS_COMPLETED {
			_, err = repository.UpdateBookingAndPaymentStatusInTxn(ctx, tx, req.BookingId, CustomUtils.GetPtr(constants.BOOKING_STATUS_COMPLETED), nil)
			if err != nil {
				return CustomUtils.Invalid("Failed to update booking status: " + err.Error())
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return updatedRide, nil
}

func getDriverIdForRideHistory(req dto.UpdateRideRequest) *string {
	var driverId *string = nil

	switch req.PreviousStatus {
	case constants.RIDE_STATUS_GMC_REQUESTED, constants.RIDE_STATUS_DROP_ASSIGNED, constants.RIDE_STATUS_DROP_ENROUTE, constants.RIDE_STATUS_DROP_COMPLETED:
		driverId = req.DropDriverId
	case constants.RIDE_STATUS_REQUESTED, constants.RIDE_STATUS_PICK_ASSIGNED, constants.RIDE_STATUS_PICK_COMPLETED, constants.RIDE_STATUS_PARKED:
		driverId = req.PickDriverId
	}
	return driverId
}

//		RequestGMCService called user wants to get the car back. The following takes place in this transaction:
//
//		1. get ride details from ride id and check the current ride status is same as PARKED
//	 	2. get the booking details, and validate payment related logics
//		3. update the ride_status, iff the payment_status is NOT_REQUIRED or PAID
//		4. add a new entry in ride_history table with the updated status
func RequestGMCService(ctx context.Context, pool *pgxpool.Pool, req dto.UpdateRideRequest, currentUser middleware.AuthUser) (*models.Ride, error) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updatedRide *models.Ride

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// ------------Step 1
		ride, err := repository.GetRideDetailsByID(ctx, tx, req.RideID, database.ForUpdate)
		if err != nil {
			return err
		}

		if ride.RideStatus != constants.RIDE_STATUS_PARKED {
			return CustomUtils.Invalid("Invalid operation when the current ride status is: " + ride.RideStatus)
		}

		// -------------Step 2
		booking, err := repository.GetBookingByIdInTxn(ctx, tx, req.BookingId, database.ForShare)
		if err != nil {
			return err
		}
		paymentStatus := booking.PaymentStatus

		// --------------Step 3
		if CustomUtils.Deref(paymentStatus) == constants.PAYMENT_STATUS_NOT_REQUIRED || CustomUtils.Deref(paymentStatus) == constants.PAYMENT_STATUS_PAID {
			updatedRide, err = repository.StepUpRideInTxn(ctx, tx, req)
			if err != nil {
				// if errors.Is(err, pgx.ErrNoRows) {
				// 	return CustomUtils.Invalid("No ride found with given ride_id")
				// }
				return err
			}

			// -----------Step 4
			r := buildRideHistoryReq(req.RideID, &req.PreviousStatus, updatedRide.RideStatus, &currentUser.ID, &currentUser.Role, nil)
			_, err = repository.AddRideHistoryInTxn(ctx, tx, r)
			if err != nil {
				return err
			}
			return nil
		}
		return CustomUtils.Invalid("Cannot request this vehicle back the booking status is " + CustomUtils.Deref(paymentStatus))
	})
	if err != nil {
		return nil, err
	}

	return updatedRide, nil
}
