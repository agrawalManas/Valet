package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/models"
	"context"

	"github.com/jackc/pgx/v5"
)

const rideColumns = `booking_id, ride_status, created_by, pick_driver_id, drop_driver_id, supervisor_id, pick_review_id, drop_review_id
					 ride_review_id, parking_location, parked_car_images, otp`
const rideHistoryColumns = `ride_id, from_status, to_status, actor_id, actor_role, driver_id`

func scanRide(row pgx.Row) (*models.Ride, error) {
	var ride models.Ride

	err := row.Scan(&ride.Id, &ride.BookingId, &ride.RideStatus, &ride.CreatedBy, &ride.PickReviewId, &ride.DropDriverId, &ride.PickReviewId, &ride.DropReviewId,
		&ride.RideReviewId, &ride.ParkingLocation, &ride.Otp, &ride.CreatedAt, &ride.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ride, nil
}

func scanRideHistory(row pgx.Row) (*models.RideHistory, error) {
	var rideHistory models.RideHistory

	err := row.Scan(&rideHistory.RideId, &rideHistory.PreviousStatus, &rideHistory.NewStatus, &rideHistory.ActorId, &rideHistory.ActorRole,
		&rideHistory.DriverId, &rideHistory.CreatedAt, &rideHistory.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rideHistory, nil
}

func CreateNewRideInTxn(ctx context.Context, db database.DBTX, bookingId, createdBy string, otp int, supervisorId, pickDriverId *string) (*models.Ride, error) {
	query := `INSERT INTO rides (booking_id, created_by, otp)
			  VALUES ($1, $2, $3)
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, bookingId, createdBy, otp)
	return scanRide(row)
}

func AddRideHistoryInTxn(ctx context.Context, db database.DBTX, rideHistory *models.RideHistory) (*models.RideHistory, error) {
	query := `INSERT INTO ride_history (` + rideHistoryColumns + `)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING ` + rideHistoryColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, rideHistory.RideId, rideHistory.PreviousStatus, rideHistory.NewStatus, rideHistory.ActorRole, rideHistory.DriverId)
	return scanRideHistory(row)
}
