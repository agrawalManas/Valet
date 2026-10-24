package repository

import (
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, mobileNumber string, role string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = ` INSERT INTO users(mobile_number, role)
				  VALUES ($1, $2)
				  RETURNING id, full_name, email, mobile_number, role, is_active, created_at, updated_at
	`

	var createdUser models.User

	err := pool.QueryRow(ctx, query, mobileNumber, role).Scan(
		&createdUser.Id,
		&createdUser.FullName,
		&createdUser.Email,
		&createdUser.MobileNumber,
		&createdUser.Role,
		&createdUser.IsActive,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &createdUser, nil
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

	var user models.User
	err := pool.QueryRow(ctx, query, mobileNumber).Scan(
		&user.Id,
		&user.FullName,
		&user.Email,
		&user.MobileNumber,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(pool *pgxpool.Pool, id string, fullName *string, email *string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` UPDATE users 
			   SET full_name = COALESCE($1, full_name),
			   	   email = COALESCE($2, email),
			       updated_at = CURRENT_TIMESTAMP
			   WHERE id = $3
			   RETURNING id, full_name, email, mobile_number, role, is_active, created_at, updated_at
	`
	var user models.User

	err := pool.QueryRow(ctx, query, fullName, email, id).Scan(
		&user.Id,
		&user.FullName,
		&user.Email,
		&user.MobileNumber,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
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
			   RETURNING id, full_name, email, mobile_number, role, is_active, created_at, updated_at
			`

	var user models.User

	err := pool.QueryRow(ctx, query, targetId, role, isActive).Scan(
		&user.Id,
		&user.FullName,
		&user.Email,
		&user.MobileNumber,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
