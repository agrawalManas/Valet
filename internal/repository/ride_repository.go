package repository

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const rideColumns = `booking_id, site_id, ride_status, created_by, pick_driver_id, drop_driver_id, supervisor_id, pick_review_id, drop_review_id,
					 ride_review_id, parking_location, parked_car_images, otp`
const rideHistoryColumns = `ride_id, from_status, to_status, actor_id, actor_role, driver_id`
const rideReviewColumns = `id, ride_id, reviewer_id, rating, review, created_at, updated_at`

func scanRideReview(row pgx.Row) (*models.RideReview, error) {
	var review models.RideReview

	err := row.Scan(&review.Id, &review.RideId, &review.ReviewerId, &review.Rating, &review.Review, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func scanRide(row pgx.Row) (*models.Ride, error) {
	var ride models.Ride

	err := row.Scan(&ride.Id, &ride.BookingId, &ride.SiteId, &ride.RideStatus, &ride.CreatedBy, &ride.PickDriverId, &ride.DropDriverId, &ride.SupervisorId, &ride.PickReviewId, &ride.DropReviewId,
		&ride.RideReviewId, &ride.ParkingLocation, &ride.ParkedCarImages, &ride.Otp, &ride.CreatedAt, &ride.UpdatedAt)
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

func CreateNewRideInTxn(ctx context.Context, db database.DBTX, bookingId, siteId, createdBy string, otp int, supervisorId, pickDriverId *string) (*models.Ride, error) {
	query := `INSERT INTO rides (booking_id, site_id, created_by, otp)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, bookingId, siteId, createdBy, otp)
	return scanRide(row)
}

func GetRideDetailsByID(ctx context.Context, db database.DBTX, rideId string, lock database.LockMode) (*models.Ride, error) {

	query := `SELECT id, ` + rideColumns + `, created_at, updated_at
			  FROM rides
			  WHERE id = $1` + lock.Clause()
	row := db.QueryRow(ctx, query, rideId)

	return scanRide(row)
}

func StepUpRideInTxn(ctx context.Context, db database.DBTX, req dto.UpdateRideRequest, driverId *string) (*models.Ride, error) {
	var query string
	var row pgx.Row

	if req.PreviousStatus == constants.RIDE_STATUS_FEEDBACK {
		return markRideAsCompleted(ctx, db, req.RideID)
	}

	switch req.PreviousStatus {
	case constants.RIDE_STATUS_REQUESTED:
		query = `UPDATE rides
			  SET ride_status = $3,
			  pick_driver_id = COALESCE($2, pick_driver_id),
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, driverId, constants.RIDE_STATUS_PICK_ASSIGNED)

	case constants.RIDE_STATUS_PICK_ASSIGNED:
		query = `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, constants.RIDE_STATUS_PICK_COMPLETED)

	case constants.RIDE_STATUS_PICK_COMPLETED:
		query = `UPDATE rides
			  SET ride_status = $4,
			  parking_location = COALESCE($2, parking_location),
			  parked_car_images = COALESCE($3, parked_car_images),
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.ParkingLocation, req.ParkedCarImages, constants.RIDE_STATUS_PARKED)

	case constants.RIDE_STATUS_PARKED:
		query = `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, constants.RIDE_STATUS_GMC_REQUESTED)

	case constants.RIDE_STATUS_GMC_REQUESTED:
		query = `UPDATE rides
			  SET ride_status = $3,
			  drop_driver_id = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, driverId, constants.RIDE_STATUS_DROP_ASSIGNED)

	case constants.RIDE_STATUS_DROP_ASSIGNED:
		query = `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, constants.RIDE_STATUS_DROP_ENROUTE)

	case constants.RIDE_STATUS_DROP_ENROUTE:
		query = `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, constants.RIDE_STATUS_DROP_COMPLETED)

	case constants.RIDE_STATUS_DROP_COMPLETED:
		query = `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, constants.RIDE_STATUS_KEY_RETURNED)

	case constants.RIDE_STATUS_KEY_RETURNED:
		query = `UPDATE rides
			  SET ride_status = $3,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.RideReviewId, constants.RIDE_STATUS_FEEDBACK)
	}
	return scanRide(row)
}

func markRideAsCompleted(ctx context.Context, db database.DBTX, rideId string) (*models.Ride, error) {
	query := `UPDATE rides
			  SET ride_status = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, rideId, constants.RIDE_STATUS_COMPLETED)
	return scanRide(row)
}

func AddRideHistoryInTxn(ctx context.Context, db database.DBTX, rideHistory *models.RideHistory) (*models.RideHistory, error) {
	query := `INSERT INTO ride_history (` + rideHistoryColumns + `)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING ` + rideHistoryColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, rideHistory.RideId, rideHistory.PreviousStatus, rideHistory.NewStatus, rideHistory.ActorId, rideHistory.ActorRole, rideHistory.DriverId)
	return scanRideHistory(row)
}

func GetRideTimeline(pool *pgxpool.Pool, rideId string) ([]*models.RideHistory, error) {
	var ctx context.Context

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT ` + rideHistoryColumns + `, created_at, updated_at
			  FROM ride_history
			  WHERE ride_id = $1
			  ORDER BY created_at`

	rows, err := pool.Query(ctx, query, rideId)
	if err != nil {
		return nil, err
	}

	var timeline []*models.RideHistory
	defer rows.Close()

	for rows.Next() {
		ride, err := scanRideHistory(rows)
		if err != nil {
			return nil, err
		}
		timeline = append(timeline, ride)
	}
	return timeline, nil
}

func GetDriverRequestsInTxn(ctx context.Context, db database.DBTX, siteId, rideStatus string) ([]*models.Ride, error) {
	query := `SELECT id` + rideColumns + `, created_at, updated_at
			  FROM rides
			  WHERE site_id = $1 AND ride_status = $2
			  ORDER BY created_at`

	rows, err := db.Query(ctx, query, siteId, rideStatus)
	if err != nil {
		return nil, err
	}

	var availableRides []*models.Ride
	defer rows.Close()

	for rows.Next() {
		ride, err := scanRide(rows)
		if err != nil {
			return nil, err
		}
		availableRides = append(availableRides, ride)
	}
	return availableRides, nil
}

func GetSiteIDAgainstRideId(pool *pgxpool.Pool, rideId string) (*string, error) {
	var ctx context.Context

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var siteId string
	query := `SELECT site_id
			  FROM rides
			  WHERE ride_id = $1`

	err := pool.QueryRow(ctx, query, rideId).Scan(siteId)
	if err != nil {
		return nil, err
	}

	return &siteId, nil
}

func AddRideReviewInTx(ctx context.Context, db database.DBTX, rideId, reviewerId string, rating float32, review *string) (*models.RideReview, error) {
	query := `INSERT into ride_reviews (reviewer_id, ride_id, rating, review)
			  VALUES ($1, $2, $3, $4)
			  RETURNING ` + rideReviewColumns
	row := db.QueryRow(ctx, query, reviewerId, rideId, rating, review)

	return scanRideReview(row)
}
