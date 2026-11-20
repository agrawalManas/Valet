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

const rideColumns = `booking_id, site_id, ride_status, created_by, pick_driver_id, drop_driver_id, supervisor_id, pick_review_id, drop_review_id
					 ride_review_id, parking_location, parked_car_images, otp`
const rideHistoryColumns = `ride_id, from_status, to_status, actor_id, actor_role, driver_id`

func scanRide(row pgx.Row) (*models.Ride, error) {
	var ride models.Ride

	err := row.Scan(&ride.Id, &ride.BookingId, &ride.SiteId, &ride.RideStatus, &ride.CreatedBy, &ride.PickReviewId, &ride.DropDriverId, &ride.PickReviewId, &ride.DropReviewId,
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

func CreateNewRideInTxn(ctx context.Context, db database.DBTX, bookingId, siteId, createdBy string, otp int, supervisorId, pickDriverId *string) (*models.Ride, error) {
	query := `INSERT INTO rides (booking_id, site_id, created_by, otp)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, bookingId, siteId, createdBy, otp)
	return scanRide(row)
}

func GetRideDetailsByID(ctx context.Context, db database.DBTX, rideId string) (*models.Ride, error) {

	query := `SELECT ` + rideColumns + `
			  FROM rides
			  WHERE id = $1`
	row := db.QueryRow(ctx, query, rideId)

	return scanRide(row)
}

func StepUpRideInTxn(ctx context.Context, db database.DBTX, req dto.UpdateRideRequest) (*models.Ride, error) {
	var query string
	var row pgx.Row

	if req.PreviousStatus == constants.RIDE_STATUS_FEEDBACK {
		return markRideAsCompleted(ctx, db, req.RideID)
	}

	switch req.PreviousStatus {
	case constants.RIDE_STATUS_REQUESTED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_PICK_ASSIGNED + `,
			  pick_driver_id = COALESCE($2, pick_driver_id),
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.PickDriverId)

	case constants.RIDE_STATUS_PICK_ASSIGNED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_PICK_COMPLETED + `,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID)

	case constants.RIDE_STATUS_PICK_COMPLETED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_PARKED + `,
			  parking_location = COALESCE($2, parking_location),
			  parked_car_images = COALESCE($3, parked_car_images),
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.ParkingLocation, req.ParkedCarImages)

	case constants.RIDE_STATUS_PARKED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_GMC_REQUESTED + `
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID)

	case constants.RIDE_STATUS_GMC_REQUESTED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_DROP_ASSIGNED + `,
			  drop_driver_id = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.DropDriverId)

	case constants.RIDE_STATUS_DROP_ASSIGNED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_DROP_ENROUTE + `,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID)

	case constants.RIDE_STATUS_DROP_ENROUTE:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_DROP_COMPLETED + `,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID)

	case constants.RIDE_STATUS_DROP_COMPLETED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_KEY_RETURNED + `,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID)

	case constants.RIDE_STATUS_KEY_RETURNED:
		query = `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_FEEDBACK + `,
			  ride_review_id = $2,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`
		row = db.QueryRow(ctx, query, req.RideID, req.RideReviewId)
	}
	return scanRide(row)
}

func markRideAsCompleted(ctx context.Context, db database.DBTX, rideId string) (*models.Ride, error) {
	query := `UPDATE rides
			  SET ride_status = ` + constants.RIDE_STATUS_COMPLETED + `,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, ` + rideColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, rideId)
	return scanRide(row)
}

func AddRideHistoryInTxn(ctx context.Context, db database.DBTX, rideHistory *models.RideHistory) (*models.RideHistory, error) {
	query := `INSERT INTO ride_history (` + rideHistoryColumns + `)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING ` + rideHistoryColumns + `, created_at, updated_at`

	row := db.QueryRow(ctx, query, rideHistory.RideId, rideHistory.PreviousStatus, rideHistory.NewStatus, rideHistory.ActorRole, rideHistory.DriverId)
	return scanRideHistory(row)
}

func GetRideTimeline(pool *pgxpool.Pool, rideId string) ([]*models.RideHistory, error) {
	var ctx context.Context

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT ` + rideHistoryColumns + `
			  FROM ride_history
			  WHERE ride_id = $1
			  ORDER BY created_at`

	rows, err := pool.Query(ctx, query, rideId)
	if err != nil {
		return nil, err
	}

	var timeline []*models.RideHistory

	for rows.Next() {
		ride, err := scanRideHistory(rows)
		if err != nil {
			return nil, err
		}
		timeline = append(timeline, ride)
	}
	return timeline, nil
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
