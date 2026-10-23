package repository

import (
	"agrawalManas/Valet/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, mobileNumber string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = ` INSERT INTO users(mobile_number)
				  VALUES ($1)
				  RETURNING id, full_name, email, mobile_number, role, created_at, updated_at
	`

	var createdUser models.User

	err := pool.QueryRow(ctx, query, mobileNumber).Scan(
		&createdUser.Id,
		&createdUser.FullName,
		&createdUser.Email,
		&createdUser.MobileNumber,
		&createdUser.Role,
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
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(pool *pgxpool.Pool, fullName *string, email *string, role *string, id string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := ` UPDATE users 
			   SET full_name = COALESCE($1, full_name),
			   	   email = COALESCE($2, email),
			       role = COALESCE($3, role),
			       updated_at = CURRENT_TIMESTAMP
			   WHERE id = $4
			   RETURNING id, full_name, email, mobile_number, role, created_at, updated_at
	`
	var user models.User

	err := pool.QueryRow(ctx, query, fullName, email, role, id).Scan(
		&user.Id,
		&user.FullName,
		&user.Email,
		&user.MobileNumber,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
