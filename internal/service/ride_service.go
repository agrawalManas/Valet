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

//	StartRideService runs upon the assumption that all the site/slot related validations were done while creating the booking,
//	allowing us to start the ride cycle. The following takes place in this transaction:
//
//	1. get booking details from booking id and check if the booking date is of today, the vehicle number matches and booking status is Scheduled
//	2. resolve the current user id and
//	3. update booking status to active in bookings table against the booking id
//	4. create a new ride entry with default ride status
//	5. create a new entry in ride history table and then add another entry if the current user is driver
//		as for the driver created rides, the pickup_driver is caller only
//
// If any step fails, nothing is written: no orphan entry in the rides and ride_history table.
func StartRideService(ctx context.Context, pool *pgxpool.Pool, bookingId, vrn string, currentUser middleware.AuthUser) (*models.Ride, error) {
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

		driverId, supervisorId, err = resolveDriverSupervisorId(ctx, tx, currentUser.ID, currentUser.Role, siteId)
		if err != nil {
			return err
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

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ride, nil
}

func validateSupervisorForSite(ctx context.Context, tx database.DBTX, siteId string, supervisorId string) error {
	exists, err := repository.CheckSupervisorAgainstASiteInTxn(ctx, tx, siteId, supervisorId)
	if err != nil {
		return err
	}

	if !exists {
		return CustomUtils.Invalid("This supervisor does not belong to this site")
	}

	return nil
}

func resolveDriverSupervisorId(ctx context.Context, tx database.DBTX, currentUserId, currentUserRole, siteId string) (driverId, supervisorId *string, err error) {
	var exists bool

	switch currentUserRole {
	case constants.ROLE_DRIVER:
		exists, err = repository.CheckDriverAgainstSiteIdInTx(ctx, tx, currentUserId, siteId)
		if err != nil {
			return nil, nil, err
		}

		if !exists {
			return nil, nil, CustomUtils.Invalid("This driver doest not belong to this site")
		}
		driverId = &currentUserId
	case constants.ROLE_SUPERVISOR:
		err = validateSupervisorForSite(ctx, tx, siteId, currentUserId)
		if err != nil {
			return nil, nil, err
		}

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

	if booking.Status == constants.BOOKING_STATUS_ACTIVE {
		return nil, CustomUtils.Invalid("This ride is already active, cannot start a new ride")
	}

	if booking.Status == constants.BOOKING_STATUS_COMPLETED {
		return nil, CustomUtils.Invalid("This ride is already completed")
	}

	if booking.Status != constants.BOOKING_STATUS_SCHEDULED {
		return nil, CustomUtils.Invalid("Cannot start a new ride, as the current booking is " + booking.Status)
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
//	2. if it is pick/drop request, check whether the driver is associated with site or not then update the current entry
//	   also check if the provided supervisor correspond to this site when current ride status is drop_complete
//	3. create a new ride entry with the new ride statuses
//	4. update booking status to complete if applicable
//
// If any step fails, nothing is written: no orphan entry in the rides and ride_history table.
func UpdateRideService(ctx context.Context, pool *pgxpool.Pool, req dto.UpdateRideRequest, currentUser middleware.AuthUser) (*models.Ride, error) {
	var err error
	var updatedRide *models.Ride

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	var driverId *string
	previousStatus := req.PreviousStatus
	isPickOrDropRequest := previousStatus == constants.RIDE_STATUS_REQUESTED || previousStatus == constants.RIDE_STATUS_GMC_REQUESTED
	if isPickOrDropRequest && currentUser.Role != constants.ROLE_DRIVER {
		err = CustomUtils.Invalid("This ride can be moved forward when a driver accepts the request")
		return nil, err
	}
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// -----------Step 1
		ride, err := repository.GetRideDetailsByID(ctx, tx, req.RideID, database.ForUpdate)
		if err != nil {
			return err
		}

		var status = ride.RideStatus

		if status == constants.RIDE_STATUS_COMPLETED {
			return CustomUtils.Invalid("This ride has already been completed!")
		}

		if previousStatus != status {
			return CustomUtils.Invalid("Current ride status " + status + " does not match the given previous_status")
		}

		// -----------Step 2
		if isPickOrDropRequest {
			driverId, _, err = resolveDriverSupervisorId(ctx, tx, currentUser.ID, currentUser.Role, ride.SiteId)
			if err != nil {
				return err
			}
			driverId = &currentUser.ID
		}

		supervisorId := getSupervisorId(currentUser.ID, currentUser.Role, req.SupervisorId)
		if status == constants.RIDE_STATUS_DROP_COMPLETED && supervisorId != nil {
			err = validateSupervisorForSite(ctx, tx, ride.SiteId, CustomUtils.Deref(req.SupervisorId))
			if err != nil {
				return CustomUtils.Invalid("The given supervisor id is not associated with this site")
			}
		}

		updatedRide, err = repository.StepUpRideInTxn(ctx, tx, req, driverId)
		if err != nil {
			// if errors.Is(err, pgx.ErrNoRows) {
			// 	return CustomUtils.Invalid("No ride found with given ride_id")
			// }
			return err
		}

		// ------------Step 3
		r := buildRideHistoryReq(req.RideID, &req.PreviousStatus, updatedRide.RideStatus, &currentUser.ID, &currentUser.Role, getDriverIdForRideHistory(req, currentUser.ID))
		_, err = repository.AddRideHistoryInTxn(ctx, tx, r)
		if err != nil {
			return err
		}

		// --------------Step 4
		if updatedRide.RideStatus == constants.RIDE_STATUS_KEY_RETURNED {
			_, err = repository.UpdateBookingAndPaymentStatusInTxn(ctx, tx, ride.BookingId, CustomUtils.GetPtr(constants.BOOKING_STATUS_COMPLETED), nil)
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

func getSupervisorId(currentUserID, currentUserRole string, supervisorId *string) (id *string) {
	if currentUserRole == constants.ROLE_SUPERVISOR {
		return &currentUserID
	}
	return supervisorId
}

func getDriverIdForRideHistory(req dto.UpdateRideRequest, currentUserId string) *string {
	var driverId *string = nil

	switch req.PreviousStatus {
	case constants.RIDE_STATUS_GMC_REQUESTED, constants.RIDE_STATUS_REQUESTED:
		driverId = &currentUserId
	}
	return driverId
}

//		RequestGMCService should be called user wants to get the car back. The following takes place in this transaction:
//
//		1. get ride details from ride id and check the current ride status is same as PARKED
//	 	2. get the booking details, and validate payment related logics
//		3. update the ride_status, iff the payment_status is NOT_REQUIRED or PAID
//		4. add a new entry in ride_history table with the updated status
func RequestGMCService(pool *pgxpool.Pool, req dto.UpdateRideRequest, currentUser middleware.AuthUser) (*models.Ride, error) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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

		if req.BookingId != ride.BookingId {
			return CustomUtils.Invalid("The given booking is not associated with the ride id")
		}

		// -------------Step 2
		booking, err := repository.GetBookingByIdInTxn(ctx, tx, ride.BookingId, database.ForShare)
		if err != nil {
			return err
		}

		if booking.UserId != currentUser.ID {
			return CustomUtils.Invalid("This booking is not applicable to this user")
		}
		paymentStatus := booking.PaymentStatus

		// --------------Step 3
		if CustomUtils.Deref(paymentStatus) == constants.PAYMENT_STATUS_NOT_REQUIRED || CustomUtils.Deref(paymentStatus) == constants.PAYMENT_STATUS_PAID {
			updatedRide, err = repository.StepUpRideInTxn(ctx, tx, req, nil)
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
		return CustomUtils.Invalid("Cannot request this vehicle back when the payment status is " + CustomUtils.Deref(paymentStatus))
	})
	if err != nil {
		return nil, err
	}

	return updatedRide, nil
}

// AddRideReview allows user to add a review for a ride, in single transaction involving following steps:
//
// 1. get ride details from ride id and check the current ride status is same as RIDE_FEEDBACK
// 2. create a new entry in ride_reviews table
// 3. update the ride_status for this ride to COMPLETED
// 4. add a new entry in ride_history table with the updated status
func AddRideReview(pool *pgxpool.Pool, req dto.AddRideReviewDTO, currentUser middleware.AuthUser) (rideReview *models.RideReview, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// ----------Step 1
		ride, err := repository.GetRideDetailsByID(ctx, tx, req.RideId, database.ForUpdate)
		if err != nil {
			return err
		}

		err = resolveRideStatusForFeedback(ride.RideStatus)
		if err != nil {
			return err
		}

		// -----------Step 2
		rideReview, err = repository.AddRideReviewInTx(ctx, tx, req.RideId, currentUser.ID, req.Rating, req.Review)
		if err != nil {
			return err
		}

		// -----------Step 3
		req := dto.UpdateRideRequest{
			RideID:         ride.Id,
			BookingId:      ride.BookingId,
			PreviousStatus: constants.RIDE_STATUS_FEEDBACK,
			RideReviewId:   &rideReview.Id,
		}

		updatedRide, err := repository.StepUpRideInTxn(ctx, tx, req, nil)
		if err != nil {
			return err
		}

		// -------------Step 4
		r := buildRideHistoryReq(updatedRide.Id, CustomUtils.GetPtr(constants.RIDE_STATUS_FEEDBACK), constants.RIDE_STATUS_COMPLETED, &currentUser.ID, &currentUser.Role, nil)
		_, err = repository.AddRideHistoryInTxn(ctx, tx, r)
		if err != nil {
			return err
		}

		return nil
	})

	return rideReview, err
}

func resolveRideStatusForFeedback(currentStatus string) error {
	if currentStatus == constants.RIDE_STATUS_FEEDBACK {
		return nil
	} else if currentStatus == constants.RIDE_STATUS_COMPLETED {
		return CustomUtils.Invalid("This ride has already been completed cannot add a feedback")
	}
	return CustomUtils.Invalid("Cannot add feedback when the current status is: " + currentStatus)
}
