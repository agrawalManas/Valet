package repository

import (
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const slotColumns = "id, start_time, end_time, site_id, is_open, created_at, updated_at"

func scanSlot(row pgx.Row) (*models.Slot, error) {
	var slot models.Slot

	err := row.Scan(&slot.Id, &slot.StartTime, &slot.EndTime, &slot.SiteId, &slot.IsOpen, &slot.CreatedAt, &slot.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &slot, nil
}

func CreateSlot(pool *pgxpool.Pool, startTime, endTime string, siteId string, isOpen bool) (*models.Slot, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const query = `INSERT INTO slots(start_time, end_time, site_id, is_open)
				   VALUES ($1, $2, $3, $4)
				   RETURNING ` + slotColumns

	row := pool.QueryRow(ctx, query, startTime, endTime, siteId, isOpen)
	return scanSlot(row)
}

func UpdateSlot(pool *pgxpool.Pool, slotId string, startTime, endTime *string, isOpen *bool) (*models.Slot, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` UPDATE slots 
			   SET start_time = COALESCE($2, start_time),
				   end_time = COALESCE($3, end_time),
			       is_open = COALESCE($4, is_open),
				   updated_at = CURRENT_TIMESTAMP
			   WHERE id = $1
			   RETURNING ` + slotColumns

	row := pool.QueryRow(ctx, query, slotId, startTime, endTime, isOpen)
	return scanSlot(row)
}
