package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const bookingColumns = `user_id, created_by, site_id, slot_id, vehicle_number, vehicle_name, for_date, status, booking_type, booking_mode,
						payment_method, payment_status, cancelled_at, cancel_reason, valet_amount, insurance_opted, insurance_amount, taxes`

func scanBooking(row pgx.Row) (*models.Booking, error) {
	var booking models.Booking
	err := row.Scan(&booking.Id, &booking.UserId, &booking.CreatedBy, &booking.SiteId, &booking.SlotId, &booking.VehicleNumber, &booking.VehicleName,
		&booking.BookingDate, &booking.Status, &booking.BookingType, &booking.BookingMode, &booking.PaymentMethod, &booking.PaymentStatus,
		&booking.CancelledAt, &booking.CancelReason, &booking.ValetAmount, &booking.InsuranceOpted, &booking.InsuranceAmount,
		&booking.TotalTax, &booking.TotalAmount, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func GetBookingCountForASite(pool *pgxpool.Pool, siteId string, bookingDate time.Time) (int, error) {
	return 0, nil
}

func GetBookingCountForASiteInTransaction(ctx context.Context, db database.DBTX, siteId string, bookingDate time.Time) (*int, error) {
	var bookingCount int

	query := `
			 SELECT COUNT(*)
			 FROM bookings
			 WHERE site_id = $1 AND for_date = $2
	`

	err := db.QueryRow(ctx, query, siteId, bookingDate.Format("2006-01-02")).Scan(&bookingCount)
	if err != nil {
		return nil, err
	}
	return &bookingCount, err
}

func GetActiveScheduledBookingsForVRN(ctx context.Context, db database.DBTX, siteId, vrn string) (*int, error) {
	var bookingCount int

	query := `
			 SELECT COUNT(*)
			 FROM bookings
			 WHERE site_id = $1 AND vehicle_number = $2 AND status IN ('SCHEDULED', 'ACTIVE')
	`

	err := db.QueryRow(ctx, query, siteId, vrn).Scan(&bookingCount)
	if err != nil {
		return nil, err
	}
	return &bookingCount, err
}

func CreateBookingInTxn(ctx context.Context, db database.DBTX, userId, createdBy, paymentStatus string, bookingReq *models.Booking) (*models.Booking, error) {

	query := `
		INSERT INTO bookings (` + bookingColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING ` + `id, ` + bookingColumns + ", total_amount, created_at, updated_at"

	row := db.QueryRow(ctx, query, userId, createdBy, bookingReq.SiteId, bookingReq.SlotId, bookingReq.VehicleNumber, bookingReq.VehicleName,
		bookingReq.BookingDate, "SCHEDULED", bookingReq.BookingType, bookingReq.BookingMode, bookingReq.PaymentMethod, paymentStatus,
		nil, nil, bookingReq.ValetAmount, bookingReq.InsuranceOpted, bookingReq.InsuranceAmount, bookingReq.TotalTax)

	return scanBooking(row)
}

func UpdateBookingAndPaymentStatusInTxn(ctx context.Context, db database.DBTX, bookingId string, status, paymentStatus *string) (*models.Booking, error) {
	query := `UPDATE bookings
			  SET status = COALESCE($2, status), payment_status = COALESCE($3, payment_status), updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING ` + `id, ` + bookingColumns + ", total_amount, created_at, updated_at"

	row := db.QueryRow(ctx, query, bookingId, status, paymentStatus)
	return scanBooking(row)
}
