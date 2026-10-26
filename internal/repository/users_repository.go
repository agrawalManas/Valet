package repository

import (
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const userColumns = `id::text, full_name, email, mobile_number, role, is_active, gender, dob::text, created_at, updated_at`

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.Id, &u.FullName, &u.Email, &u.MobileNumber, &u.Role, &u.IsActive, &u.Gender, &u.Dob, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateUser(pool *pgxpool.Pool, mobileNumber string, role string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = ` INSERT INTO users(mobile_number, role)
				  VALUES ($1, $2)
				  RETURNING ` + userColumns

	row := pool.QueryRow(ctx, query, mobileNumber, role)
	return scanUser(row)
}

func GetUserByMobile(pool *pgxpool.Pool, mobileNumber string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = `SELECT * 
				 FROM users
				 WHERE mobile_number = $1
	`

	row := pool.QueryRow(ctx, query, mobileNumber)
	return scanUser(row)
}

func UpdateUser(pool *pgxpool.Pool, id string, fullName *string, email *string, dob *string, gender *string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` UPDATE users 
			   SET full_name = COALESCE($2, full_name),
			   	   email = COALESCE($3, email),
			   	   gender = COALESCE($4, gender),
			   	   dob = COALESCE($5, dob),
			       updated_at = CURRENT_TIMESTAMP
			   WHERE id = $1
			   RETURNING ` + userColumns

	row := pool.QueryRow(ctx, query, id, fullName, email, gender, dob)
	return scanUser(row)
}

func UpdateUserAccess(pool *pgxpool.Pool, targetId string, role *string, isActive *bool) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` UPDATE users 
			   SET role = COALESCE($2, role),
				   is_active = COALESCE($3, is_active),
			       updated_at = CURRENT_TIMESTAMP
			   WHERE id = $1
			   RETURNING ` + userColumns

	row := pool.QueryRow(ctx, query, targetId, role, isActive)
	return scanUser(row)
}

// FindUserByMobileForUpdate returns the user and locks the row until the
// transaction ends, so the user cannot be demoted or deactivated mid-flow.
// Returns pgx.ErrNoRows when no user has this number.
func FindUserByMobileForUpdate(ctx context.Context, db database.DBTX, mobile string) (*models.User, error) {
	row := db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE mobile_number = $1 FOR UPDATE`,
		mobile,
	)
	return scanUser(row)
}

// CreateUserInTransaction creates a new user with the given role.
// Returns pgx.ErrNoRows when another transaction created the same mobile number
// first (ON CONFLICT DO NOTHING returns no row); the caller should re-select.
func CreateUserInTransaction(ctx context.Context, db database.DBTX, name string, email *string, mobileNumber, role string, gender string, dob time.Time) (*models.User, error) {
	fmt.Printf("date is %v", dob)
	row := db.QueryRow(ctx, `
		INSERT INTO users (full_name, email, mobile_number, role, gender, dob)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (mobile_number) DO NOTHING
		RETURNING `+userColumns,
		name, email, mobileNumber, role, gender, dob,
	)
	return scanUser(row)
}

// FillUserProfileIfMissing writes name, dob and gender only where the
// stored value is NULL. It never overwrites data that is already on file.
// The refreshed values are written back into u.
func FillUserProfileIfMissing(ctx context.Context, db database.DBTX, u *models.User, fullName string, email *string, dob time.Time, gender string) error {
	return db.QueryRow(ctx, `
		UPDATE users
		SET full_name  = COALESCE(full_name, $2),
		    email        = COALESCE(email, $3),
		    dob        = COALESCE(dob, $4),
		    gender     = COALESCE(gender, $5),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING full_name, dob::text, gender`,
		u.Id, fullName, email, dob, gender,
	).Scan(&u.FullName, &u.Dob, &u.Gender)
}
