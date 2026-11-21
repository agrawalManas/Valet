package service

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpdatePaymentStatusService handles updating the payment status for a given booking id(marking as PAID manually) in a single transaction with following steps:
//
// 1. get the ride details check the current ride_status and match the booking id
// 2. get the booking details and check the booking_status and payment_status
// 3. update the booking status, payment_status in bookings table
func UpdatePaymentStatusService(ctx context.Context, pool *pgxpool.Pool, rideId string) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var err error

		// ---------Step 1
		ride, err := repository.GetRideDetailsByID(ctx, tx, rideId, database.ForShare)
		if err != nil {
			return err
		}

		err = resolveRideStatus(ride.RideStatus)
		if err != nil {
			return CustomUtils.Invalid(err.Error())
		}

		// -----------Step 2
		booking, err := repository.GetBookingByIdInTxn(ctx, tx, ride.BookingId, database.ForUpdate)
		if err != nil {
			return err
		}

		err = resolveBookingStatus(booking.Status)
		if err != nil {
			return CustomUtils.Invalid(err.Error())
		}

		newPaymentStatus, err := resolvePaymentStatus(CustomUtils.Deref(booking.PaymentStatus))
		if err != nil || newPaymentStatus != "" {
			return CustomUtils.Invalid(err.Error())
		}

		// -----------Step 3
		_, err = repository.UpdateBookingAndPaymentStatusInTxn(ctx, tx, ride.BookingId, &booking.Status, &newPaymentStatus)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func resolveRideStatus(rideStatus string) error {
	if rideStatus == constants.RIDE_STATUS_PARKED {
		return nil
	}
	return CustomUtils.Invalid("Cannot update the current ride status: " + rideStatus)
}

func resolveBookingStatus(bookingStatus string) error {
	if bookingStatus == constants.BOOKING_STATUS_ACTIVE || bookingStatus == constants.BOOKING_STATUS_SCHEDULED {
		return nil
	}
	return CustomUtils.Invalid("Cannot update the booking status from " + bookingStatus + " to " + constants.BOOKING_STATUS_COMPLETED)
}

func resolvePaymentStatus(paymentStatus string) (string, error) {
	if paymentStatus == constants.PAYMENT_STATUS_PENDING {
		return constants.PAYMENT_STATUS_PAID, nil
	}
	return "", CustomUtils.Invalid("Cannot update the payment status for this booking: " + paymentStatus + " to " + constants.PAYMENT_STATUS_PAID)
}
