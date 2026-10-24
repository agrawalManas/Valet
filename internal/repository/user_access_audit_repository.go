package repository

import (
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func AddToUserAccessAudit(pool *pgxpool.Pool, userId string, changedBy string, oldRole *string, newRole *string, oldIsActive *bool, newIsActive bool) (*models.UserAccessAudit, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` INSERT INTO user_access_audit (user_id, changed_by, old_role, new_role, old_is_active, new_is_active)
			   VALUES ($1, $2, $3, $4, $5, $6)
			   RETURNING id, user_id, changed_by, old_role, new_role, old_is_active, new_is_active, created_at
			`

	var entry models.UserAccessAudit

	err := pool.QueryRow(ctx, query, userId, changedBy, oldRole, newRole, oldIsActive, newIsActive).Scan(
		&entry.Id,
		&entry.UserId,
		&entry.ChangedBy,
		&entry.OldRole,
		&entry.NewRole,
		&entry.OldIsActive,
		&entry.NewIsActive,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}
